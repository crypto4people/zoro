package db

import (
	"zoro/service/db/model"

	"github.com/objectbox/objectbox-go/objectbox"
)

type Database struct {
	ob            *objectbox.ObjectBox
	TxBox         *model.TxBox
	ServerBox     *model.ServerBox
	ServerInfoBox *model.ServerInfoBox
	ServerRateBox *model.ServerRateBox
	ServerSubsBox *model.ServerSubsBox
}

func New() (*Database, error) {
	ob, err := objectbox.NewBuilder().Model(model.ObjectBoxModel()).Build()
	if err != nil {
		return nil, err
	}

	return &Database{
		ob:            ob,
		TxBox:         model.BoxForTx(ob),
		ServerBox:     model.BoxForServer(ob),
		ServerInfoBox: model.BoxForServerInfo(ob),
		ServerRateBox: model.BoxForServerRate(ob),
		ServerSubsBox: model.BoxForServerSubs(ob),
	}, nil
}

func (db *Database) Close() {
	db.ob.Close()
}
