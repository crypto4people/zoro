package model

import "time"

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type Server struct {
	Id        uint64
	URL       string
	CreatedAt time.Time `objectbox:"date-nano"`
}
