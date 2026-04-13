package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var cli struct {
	Login         LoginCmd         `cmd:"" name:"login"           help:"Authenticate and write credentials to ~/.taskmaster/credentials.yml."`
	CreateTask    CreateTaskCmd    `cmd:"" name:"create-task"     help:"Create a new task."`
	GetTasks      GetTasksCmd      `cmd:"" name:"get-tasks"       help:"List tasks by status."`
	DescribeTask  DescribeTaskCmd  `cmd:"" name:"describe-task"   help:"Describe a task with blockers and addendums."`
	SetTaskStatus SetTaskStatusCmd `cmd:"" name:"set-task-status" help:"Update a task's status."`
	AddBlocker    AddBlockerCmd    `cmd:"" name:"add-blocker"     help:"Mark a task as blocked by another."`
	RemoveBlocker RemoveBlockerCmd `cmd:"" name:"remove-blocker"  help:"Remove a blocker from a task."`
	GetBlockers   GetBlockersCmd   `cmd:"" name:"get-blockers"    help:"List blockers for a task."`
	AddAddendum   AddAddendumCmd   `cmd:"" name:"add-addendum"    help:"Append an addendum to a task."`
}

func main() {
	creds, err := loadCredentials()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := kong.Parse(&cli)
	if err := ctx.Run(creds); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
