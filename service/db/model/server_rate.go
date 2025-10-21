package model

import "time"

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type ServerRate struct {
	Id       uint64
	URL      string
	Sub      string
	Div      string
	Mul      string
	Add      string
	Deadline time.Time `objectbox:"date-nano"`
}
