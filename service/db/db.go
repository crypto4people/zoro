//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen -byValue

package db

import (
	"time"
)

type TxStatus int32

const (
	TxStatusFail TxStatus = -1
	TxStatusPend TxStatus = 0
	TxStatusDone TxStatus = 1
)

type Setting struct {
	ID    uint64 `objectbox:"id"`
	Key   string `objectbox:"unique"`
	Value string
}

type Server struct {
	ID       uint64 `objectbox:"id"`
	URL      string `objectbox:"unique"`
	Version  string
	Uptime   int64
	Conns    int64
	Subs     int64
	Height   int64
	Account  []byte
	GasPrice string
	Balance  string
	RateFeed []byte
	RateSub  string
	RateDiv  string
	RateMul  string
	RateAdd  string
	Latency  int64
	Time     time.Time `objectbox:"date-nano"`
	MySubs   []Sub     `objectbox:"link"`
}

type Sub struct {
	ID     uint64 `objectbox:"id"`
	Server string `objectbox:"index"`
	Expire int    `objectbox:"index"`
	Remain int    `objectbox:"index"`
	Tx     string `objectbox:"unique"`
}

type Tx struct {
	ID     uint64 `objectbox:"id"`
	Server string
	Status TxStatus  `objectbox:"index"`
	Hash   string    `objectbox:"unique"`
	Time   time.Time `objectbox:"date-nano"`
}
