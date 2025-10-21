package db

import (
	"zoro/service/db/model"

	"github.com/objectbox/objectbox-go/objectbox"
)

type Database struct {
	ob            *objectbox.ObjectBox
	txBox         *model.TxBox
	serverBox     *model.ServerBox
	serverInfoBox *model.ServerInfoBox
	serverRateBox *model.ServerRateBox
	serverSubsBox *model.ServerSubsBox
}

func New() (*Database, error) {
	ob, err := objectbox.NewBuilder().Model(model.ObjectBoxModel()).Build()
	if err != nil {
		return nil, err
	}

	return &Database{
		ob:            ob,
		txBox:         model.BoxForTx(ob),
		serverBox:     model.BoxForServer(ob),
		serverInfoBox: model.BoxForServerInfo(ob),
		serverRateBox: model.BoxForServerRate(ob),
		serverSubsBox: model.BoxForServerSubs(ob),
	}, nil
}

func (db *Database) Close() {
	db.ob.Close()
}
