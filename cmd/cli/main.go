package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	NO_COMMAND        = "no-command"
	DEFAULT_ADDRESS   = "localhost:6100"
	SECURE_CONNECTION = false
)

var commands = map[string]func() error{
	"put":      put,
	"get":      get,
	"describe": describe,
	"mark":     mark,
	"get-tags": getTags,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("expected subcommand")
	}

	command := commands[os.Args[1]]
	if command == nil {
		log.Fatalf("invalid command %s", os.Args[1])
	}

	if err := command(); err != nil {
		log.Fatalf("%s", fmt.Errorf("running command: %w", err).Error())
	}
}

func connectionFlags(fs *flag.FlagSet, hostname *string, secure *bool, bearer *string) {
	fs.StringVar(hostname, "addr", DEFAULT_ADDRESS, "the address to connect to")
	fs.BoolVar(secure, "secure", SECURE_CONNECTION, "set to true if we need to connect over TLS")
	fs.StringVar(bearer, "bearer", "", "set this to your user bearer token")
}

func get() error {
	getCmd := flag.NewFlagSet("", flag.ExitOnError)
	var hostname string
	var bearer string
	var secure bool
	connectionFlags(getCmd, &hostname, &secure, &bearer)
	statusId := getCmd.Uint64("status-id", 0, "0) tracking, 1) completed, 2) backlog")
	tags := getCmd.String("tags", "", "tags separated by ','. If empty all tasks will be returned")
	getCmd.Parse(os.Args[2:])

	if err := withTasksClient(hostname, secure, func(client taskspb.TasksClient) error {
		var tagsToSend []string
		if *tags != "" {
			tagsToSend = strings.Split(*tags, ",")
		}
		ctx := getContext(bearer)
		task, err := client.GetTasks(ctx, &taskspb.GetTasksRequest{
			Status: taskspb.Status(*statusId),
			Tags:   tagsToSend,
		})
		if err != nil {
			return fmt.Errorf("calling client: %w", err)
		}
		for {
			res, err := task.Recv()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return fmt.Errorf("could not receive next entry: %w", err)
			}
			jsonBytes, err := protojson.Marshal(res)
			if err != nil {
				return fmt.Errorf("converting to json: %w", err)
			}
			fmt.Println(string(jsonBytes))
		}
	}); err != nil {
		return fmt.Errorf("getting task: %w", err)
	}
	return nil
}

func put() error {
	putCmd := flag.NewFlagSet("", flag.ExitOnError)
	var hostname string
	var bearer string
	var secure bool
	connectionFlags(putCmd, &hostname, &secure, &bearer)
	putCmd.Parse(os.Args[2:])

	newTask := taskspb.Task{}

	reader := bufio.NewReader(os.Stdin)

	// Name                     string
	fmt.Print("name: \n")
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read name: %w", err)
	}
	newTask.Name = strings.Trim(input, "\n")
	// MinutesToComplete        uint64
	minutes, err := readUint64(reader, "minutes to complete")
	if err != nil {
		return fmt.Errorf("failed to read minutes to complete: %w", err)
	}
	newTask.MinutesToComplete = minutes
	// EscalationTimeUnixMillis uint64
	// Priority                 Priority
	prio, err := readUint64(reader, "priority on a scale from 0 to 4")
	if err != nil {
		return fmt.Errorf("failed to read priority: %w", err)
	}
	newTask.Priority = taskspb.Priority(prio)
	// Status                   Status
	status, err := readUint64(reader, "status from a scale of 0 to 3")
	if err != nil {
		return fmt.Errorf("failed to read status: %w", err)
	}
	newTask.Status = taskspb.Status(status)
	// Tags                     []string
	fmt.Print("a set of tags delimited  by ',': \n")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read tags: %w", err)
	}
	newTask.Tags = strings.Split(strings.Trim(input, "\n"), ",")
	// Prerequisites            []uint64
	fmt.Print("a set of task_ids delimited by ',': \n")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read prereqs: %w", err)
	}
	if strings.Trim(input, "\n") != "" {
		s := strings.Split(strings.Trim(input, "\n"), ",")
		fmt.Println(strings.Trim(input, "\n"))
		p := make([]uint64, len(s))
		for i, t := range s {
			num, err := strconv.Atoi(t)
			if err != nil {
				return fmt.Errorf("parsing input prereq into number: %w", err)
			}
			p[i] = uint64(num)
		}
		newTask.Prerequisites = p
	}

	if err := withTasksClient(hostname, secure, func(client taskspb.TasksClient) error {
		resp, err := client.PutTask(getContext(bearer), &taskspb.PutTaskRequest{
			Task: &newTask,
		})
		if err != nil {
			return fmt.Errorf("putting task: %w", err)
		}
		fmt.Println(resp)
		return nil
	}); err != nil {
		return fmt.Errorf("calling tasks client for put: %w", err)
	}
	return nil
}

func describe() error {
	var hostname string
	var bearer string
	var secure bool
	descCmd := flag.NewFlagSet("", flag.ExitOnError)
	connectionFlags(descCmd, &hostname, &secure, &bearer)

	taskId := descCmd.Uint64("task-id", 0, "the ID of the task that you want to explore")
	descCmd.Parse(os.Args[2:])

	if err := withTasksClient(hostname, secure, func(client taskspb.TasksClient) error {
		stream, err := client.DescribeTask(getContext(bearer), &taskspb.DescribeTaskRequest{
			TaskId: *taskId,
		})
		if err != nil {
			return fmt.Errorf("calling task client: %w", err)
		}
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("could not receive next entry: %w", err)
			}
			jsonBytes, err := protojson.Marshal(res)
			if err != nil {
				return fmt.Errorf("converting to json: %w", err)
			}
			fmt.Println(string(jsonBytes))
		}
	}); err != nil {
		return fmt.Errorf("failed to mark task: %w", err)
	}
	return nil
}

func mark() error {
	var hostname string
	var bearer string
	var secure bool
	markCmd := flag.NewFlagSet("", flag.ExitOnError)
	connectionFlags(markCmd, &hostname, &secure, &bearer)
	markCmd.Parse(os.Args[2:])

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("taskId: \n")
	taskId, err := readUint64(reader, "task-id")
	if err != nil {
		return fmt.Errorf("failed to read taskId: %w", err)
	}

	fmt.Print("message: \n")
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}
	content := strings.Trim(input, "\n")

	if err := withTasksClient(hostname, secure, func(client taskspb.TasksClient) error {
		if _, err := client.MarkTask(getContext(bearer), &taskspb.MarkTaskRequest{
			Content: content,
			TaskId:  taskId,
		}); err != nil {
			return fmt.Errorf("calling task client: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to mark task: %w", err)
	}
	return nil
}

func getTags() error {
	var hostname string
	var bearer string
	var secure bool
	getTags := flag.NewFlagSet("", flag.ExitOnError)
	connectionFlags(getTags, &hostname, &secure, &bearer)
	getTags.Parse(os.Args[2:])

	if err := withTasksClient(hostname, secure, func(client taskspb.TasksClient) error {
		stream, err := client.GetTags(getContext(bearer), &taskspb.GetTagsRequest{})
		if err != nil {
			return fmt.Errorf("calling task client: %w", err)
		}
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("could not receive next entry: %w", err)
			}
			jsonBytes, err := protojson.Marshal(res)
			if err != nil {
				return fmt.Errorf("converting to json: %w", err)
			}
			fmt.Println(string(jsonBytes))
		}
	}); err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}
	return nil
}

func getGrpcClient(hostname string, secure bool) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if secure {
		creds = credentials.NewTLS(&tls.Config{})
	} else {
		creds = insecure.NewCredentials()
	}
	return grpc.NewClient(hostname, grpc.WithTransportCredentials(creds))
}

func withTasksClient(
	hostname string,
	secure bool,
	consumer func(client taskspb.TasksClient) error,
) error {
	conn, err := getGrpcClient(hostname, secure)
	if err != nil {
		return fmt.Errorf("connecting to grpc server: %w", err)
	}
	defer conn.Close()
	return consumer(taskspb.NewTasksClient(conn))
}

func getContext(bearer string) context.Context {
	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("Authorization", bearer),
	)
}

func readUint64(reader *bufio.Reader, desc string) (uint64, error) {
	fmt.Printf("%s :\n", desc)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read minutes to complete: %w", err)
	}
	asNumber, err := strconv.Atoi(strings.Trim(input, "\n"))
	if err != nil {
		return 0, fmt.Errorf("failed to convert minutes to complete to number: %w", err)
	}
	return uint64(asNumber), nil
}
