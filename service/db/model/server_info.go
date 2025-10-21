package model

import (
	"time"
)

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type ServerInfo struct {
	Id       uint64
	URL      string
	Conns    int64
	Chain    string
	Uptime   time.Duration `objectbox:"date-nano"`
	Height   int64
	Latency  int64
	Version  string
	Address  []byte    // cannot use common.Address due to ob
	Deadline time.Time `objectbox:"date-nano"`
}
