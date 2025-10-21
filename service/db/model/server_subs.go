package model

import (
	"encoding/json"
	"time"
)

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type ServerSubs struct {
	Id     uint64
	Tx     json.RawMessage
	Expire time.Time     `objectbox:"date-nano"`
	Remain time.Duration `objectbox:"date-nano"`
}
