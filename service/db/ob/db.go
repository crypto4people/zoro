//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen -byValue

package ob

type Account struct {
	ID      uint64 `objectbox:"id"`
	Address string `objectbox:"unique"`
	Cipher  []byte
}

type Pend struct {
	ID        uint64 `objectbox:"id"`
	Hash      string `objectbox:"unique"`
	ServerURL string
}

type Server struct {
	ID   uint64 `objectbox:"id"`
	Name string `objectbox:"unique"`
	URL  string
}

type Setting struct {
	ID    uint64 `objectbox:"id"`
	Key   string `objectbox:"unique"`
	Value string
}
