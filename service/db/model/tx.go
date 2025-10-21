package model

import (
	"encoding/json"
	"time"
)

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type Tx struct {
	Id        uint64
	Hash      []byte // common.Hash
	Server    string
	Tx        json.RawMessage
	Status    string
	CreatedAt time.Time `objectbox:"date-nano"`
}
