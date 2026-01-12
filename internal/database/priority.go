package database

type Priority int

const (
	DoBeforeSleep Status = iota
	DoImmediately
	ShouldDo
	EventuallyDo
)
