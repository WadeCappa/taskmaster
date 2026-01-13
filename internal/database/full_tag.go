package database

import (
	"time"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FullTag struct {
	id          uint64
	name        string
	timeCreated time.Time
	count       uint64
}

func (t *FullTag) ToWireType() *taskspb.GetTagsResponse {
	return &taskspb.GetTagsResponse{
		TagId:     t.id,
		Name:      t.name,
		WriteTime: timestamppb.New(t.timeCreated),
		Count:     t.count,
	}
}
