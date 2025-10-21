package model

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen
type Server struct {
	Id  uint64
	URL string
}
