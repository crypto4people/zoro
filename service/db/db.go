package db

import (
	"encoding/hex"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"
	"zoro/service/db/ob"

	"github.com/objectbox/objectbox-go/objectbox"
)

type Database struct {
	mutex    sync.RWMutex
	pends    *ob.PendBox
	servers  *ob.ServerBox
	settings *ob.SettingBox
}

func New(box *objectbox.ObjectBox) *Database {

	db := &Database{
		pends:    ob.BoxForPend(box),
		servers:  ob.BoxForServer(box),
		settings: ob.BoxForSetting(box),
	}

	return db

}

func validKey(str string) bool {
	return utf8.ValidString(str) && strings.ToLower(str) == str
}

func validURL(str string) bool {
	purl, err := url.Parse(str)
	return err == nil && purl.String() == str
}

func validHash(str string) bool {
	if len(str) != hex.EncodedLen(32) && strings.ToLower(str) != str {
		return false
	}
	data, err := hex.DecodeString(str)
	return err == nil && len(data) == 32
}
