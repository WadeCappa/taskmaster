package database

import (
	"time"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Addendum struct {
	created time.Time
	content string
}

func NewAddendum(created time.Time, content string) Addendum {
	return Addendum{
		created: created,
		content: content,
	}
}

func (a *Addendum) ToWireType() *taskspb.Addendum {
	return &taskspb.Addendum{
		Content:     a.content,
		TimeCreated: timestamppb.New(a.created),
	}
}
