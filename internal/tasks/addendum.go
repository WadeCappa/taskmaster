package tasks

import (
	"time"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type addendum struct {
	content     string
	createdTime time.Time
}

func newAddendum(
	content string,
	createdTime time.Time,
) addendum {
	return addendum{
		content:     content,
		createdTime: createdTime,
	}
}

func ToWire(addendums []addendum) []*taskspb.Addendum {
	result := make([]*taskspb.Addendum, len(addendums))
	for idx, addendum := range addendums {
		result[idx] = &taskspb.Addendum{
			Content:     addendum.content,
			CreatedTime: timestamppb.New(addendum.createdTime),
		}
	}
	return result
}
