package database

import "time"

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
