package database

import "time"

type FullTag struct {
	Id          uint64
	Name        string
	TimeCreated time.Time
	Count       uint64
}
