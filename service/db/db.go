package db

import (
	"sync"

	"github.com/crypto4people/zoro/service/db/ob"

	"github.com/objectbox/objectbox-go/objectbox"
)

type Database struct {
	mutex    sync.RWMutex
	accounts *ob.AccountBox
	pends    *ob.PendBox
	servers  *ob.ServerBox
	settings *ob.SettingBox
}

func New(box *objectbox.ObjectBox) *Database {

	db := &Database{
		accounts: ob.BoxForAccount(box),
		pends:    ob.BoxForPend(box),
		servers:  ob.BoxForServer(box),
		settings: ob.BoxForSetting(box),
	}

	return db

}
