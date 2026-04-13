package model

type Status int
const (
	Open Status = iota
	InProgress
	Closed
)
