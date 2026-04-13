package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"slices"

	authmasterpb "github.com/WadeCappa/authmaster/pkg/go/authmaster/v1"
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"github.com/alecthomas/kong"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	title1   = "Smoke task 1"
	title2   = "Smoke task 2"
	subtask  = "Smoke subtask"
	desc1    = "Created by smoke test"
	addendum = "smoke addendum"
)

var cli struct {
	Authmaster       string `help:"Authmaster server hostname." default:"localhost:50051" name:"authmaster"`
	Taskmaster       string `help:"Taskmaster server hostname." default:"localhost:6100"  name:"taskmaster"`
	AuthmasterSecure bool   `help:"Use TLS for authmaster."     default:"false"          negatable:""`
	TaskmasterSecure bool   `help:"Use TLS for taskmaster."     default:"false"          negatable:""`
}

func main() {
	kong.Parse(&cli)

	authConn, err := connect(cli.Authmaster, cli.AuthmasterSecure)
	if err != nil {
		log.Fatalf("connecting to authmaster: %v", err)
	}
	defer authConn.Close()

	taskConn, err := connect(cli.Taskmaster, cli.TaskmasterSecure)
	if err != nil {
		log.Fatalf("connecting to taskmaster: %v", err)
	}
	defer taskConn.Close()

	authClient := authmasterpb.NewAuthmasterClient(authConn)
	tasksClient := taskspb.NewTasksClient(taskConn)
	blockersClient := taskspb.NewBlockersClient(taskConn)
	addendumsClient := taskspb.NewAddendumsClient(taskConn)

	username := fmt.Sprintf("smoke-%s", uuid.NewString())
	password := fmt.Sprintf("smoke-password-%s", uuid.NewString())

	if _, err := authClient.CreateUser(context.Background(), &authmasterpb.CreateUserRequest{
		Username: username,
		Password: password,
	}); err != nil {
		log.Fatalf("could not create user: %v", err)
	}

	loginResp, err := authClient.Login(context.Background(), &authmasterpb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	token := loginResp.GetToken()

	testAuthResp, err := authClient.TestAuth(authCtx(token), &authmasterpb.TestAuthRequest{})
	if err != nil {
		log.Fatalf("could not get user-id: %v", err)
	}
	if testAuthResp.GetUserId() == 0 {
		log.Fatal("got zero'd user-id")
	}

	if err := smokeTests(authCtx(token), tasksClient, blockersClient, addendumsClient); err != nil {
		log.Fatalf("running tests: %v", err)
	}
	fmt.Println("tests pass")
}

func smokeTests(
	ctx context.Context,
	tasksClient taskspb.TasksClient,
	blockersClient taskspb.BlockersClient,
	addendumsClient taskspb.AddendumsClient,
) error {
	createResp, err := tasksClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Task: &taskspb.Task{
			Title:       title1,
			Description: desc1,
		},
	})
	if err != nil {
		return err
	}
	if createResp.GetTaskId() == 0 {
		return fmt.Errorf("got zero'd task ID")
	}
	taskId1 := createResp.GetTaskId()

	if err := verifyTaskInGetTasks(ctx, tasksClient, taskspb.Status_OPEN, taskId1); err != nil {
		return fmt.Errorf("finding task: %w", err)
	}
	descResp, err := tasksClient.DescribeTask(ctx, &taskspb.DescribeTaskRequest{TaskId: taskId1})
	if err != nil {
		return fmt.Errorf("describing task: %w", err)
	}
	t := descResp.GetTask()
	if t == nil {
		return fmt.Errorf("task should not have been nil")
	}
	if t.GetTitle() != title1 {
		return fmt.Errorf("unexpected title %q", t.GetTitle())
	}
	if t.GetStatus() != taskspb.Status_OPEN {
		return fmt.Errorf("unexpected status %v", t.GetStatus())
	}
	if _, err = tasksClient.SetTaskStatus(ctx, &taskspb.SetTaskStatusRequest{
		TaskId: taskId1,
		Status: taskspb.Status_IN_PROGRESS,
	}); err != nil {
		return fmt.Errorf("setting status: %w", err)
	}
	if err := verifyTaskInGetTasks(ctx, tasksClient, taskspb.Status_IN_PROGRESS, taskId1); err != nil {
		return fmt.Errorf("looking for task with new status: %w", err)
	}
	resp, err := tasksClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Task: &taskspb.Task{Title: title2},
	})
	if err != nil {
		return fmt.Errorf("creating second task: %w", err)
	}
	if resp.GetTaskId() == 0 {
		return fmt.Errorf("should not have received zero'd out taskId for second task")
	}
	taskId2 := resp.GetTaskId()
	if _, err := blockersClient.AddBlocker(ctx, &taskspb.AddBlockerRequest{
		TaskId:    taskId1,
		BlockedBy: taskId2,
	}); err != nil {
		return fmt.Errorf("adding blocker between tasks: %w", err)
	}
	blockers, err := blockersClient.GetBlockers(ctx, &taskspb.GetBlockersRequest{TaskId: taskId1})
	if err != nil {
		return fmt.Errorf("getting blockers for first task: %w", err)
	}
	if !slices.Contains(blockers.GetBlockedByTaskIds(), taskId2) {
		return fmt.Errorf("could not find our second task ID in our blockers list")
	}
	if _, err := blockersClient.RemoveBlocker(ctx, &taskspb.RemoveBlockerRequest{
		TaskId:    taskId1,
		BlockedBy: taskId2,
	}); err != nil {
		return fmt.Errorf("removing blocker: %w", err)
	}
	blockers, err = blockersClient.GetBlockers(ctx, &taskspb.GetBlockersRequest{TaskId: taskId1})
	if err != nil {
		return fmt.Errorf("getting blockers the second time: %w", err)
	}
	if len(blockers.GetBlockedByTaskIds()) != 0 {
		return fmt.Errorf("expected no blockers, got %v", blockers.GetBlockedByTaskIds())
	}

	if _, err := addendumsClient.AddAddendum(ctx, &taskspb.AddAddendumRequest{
		TaskId:  taskId1,
		Content: addendum,
	}); err != nil {
		return fmt.Errorf("adding addendum: %w", err)
	}
	descResp, err = tasksClient.DescribeTask(ctx, &taskspb.DescribeTaskRequest{TaskId: taskId1})
	if err != nil {
		return err
	}
	if len(descResp.GetAddendums()) != 1 {
		return fmt.Errorf("expected only one addendum")
	}
	if descResp.GetAddendums()[0].GetContent() != addendum {
		return fmt.Errorf("found incorrect addendum")
	}

	createTaskResp, err := tasksClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Task: &taskspb.Task{
			Title:        subtask,
			ParentTaskId: taskId1,
		},
	})
	if err != nil {
		return fmt.Errorf("creating subtask: %w", err)
	}
	if createTaskResp.GetTaskId() == 0 {
		return fmt.Errorf("subtask had zero'd ID")
	}
	subtaskId := createTaskResp.GetTaskId()
	subtaskDesc, err := tasksClient.DescribeTask(ctx, &taskspb.DescribeTaskRequest{TaskId: subtaskId})
	if err != nil {
		return fmt.Errorf("describing subtack: %w", err)
	}
	tsk := subtaskDesc.GetTask()
	if tsk == nil {
		return fmt.Errorf("subtask is nil")
	}
	if tsk.GetParentTaskId() != taskId1 {
		return fmt.Errorf("should have been task ID %d but was %d", taskId1, tsk.GetParentTaskId())
	}

	if _, err := tasksClient.SetTaskStatus(ctx, &taskspb.SetTaskStatusRequest{
		TaskId: taskId1,
		Status: taskspb.Status_CLOSED,
	}); err != nil {
		return fmt.Errorf("tried to close parent task: %w", err)
	}
	return nil
}

func connect(hostname string, secure bool) (*grpc.ClientConn, error) {
	var tc credentials.TransportCredentials
	if secure {
		tc = credentials.NewTLS(&tls.Config{})
	} else {
		tc = insecure.NewCredentials()
	}
	return grpc.NewClient(hostname, grpc.WithTransportCredentials(tc))
}

func authCtx(token string) context.Context {
	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("Authorization", token),
	)
}

type testCase struct {
	name string
	run  func() error
}

func verifyTaskInGetTasks(
	ctx context.Context,
	c taskspb.TasksClient,
	status taskspb.Status,
	expectedId uint64,
) error {
	stream, err := c.GetTasks(ctx, &taskspb.GetTasksRequest{Status: status})
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("task %d not found in %v stream", expectedId, status)
		}
		if err != nil {
			return err
		}
		if resp.GetTask().GetTaskId() == expectedId {
			return nil
		}
	}
}
