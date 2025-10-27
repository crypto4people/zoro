package db

import (
	"encoding/hex"
	"zoro/service/db/ob"

	"github.com/ethereum/go-ethereum/common"
)

func (db *Database) Accounts() (map[common.Address][]byte, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	accounts, err := db.accounts.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Address][]byte, len(accounts))
	for _, account := range accounts {
		result[common.HexToAddress(account.Address)] = account.Cipher
	}
	return result, nil
}

func (db *Database) DeleteAccount(address common.Address) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	account, err := db.getAccount(hex.EncodeToString(address[:]))
	if err == nil && account != nil {
		err = db.accounts.RemoveId(account.ID)
	}
	return err
}

func (db *Database) PutAccount(address common.Address, cipher []byte) error {

	addrHex := hex.EncodeToString(address[:])

	db.mutex.Lock()
	defer db.mutex.Unlock()

	account, err := db.getAccount(addrHex)
	if err != nil {
		return err
	}
	if account == nil {
		account = &ob.Account{
			Address: addrHex,
			Cipher:  cipher,
		}
	} else {
		account.Cipher = cipher
	}
	_, err = db.accounts.Put(account)
	return err
}

func (db *Database) GetAccount(address common.Address) ([]byte, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	account, err := db.getAccount(hex.EncodeToString(address[:]))
	if err != nil {
		return nil, err
	}
	return account.Cipher, nil
}

func (db *Database) getAccount(address string) (*ob.Account, error) {
	query := db.accounts.Query(ob.Account_.Address.Equals(address, true))
	defer query.Close()
	accounts, err := query.Limit(1).Find()
	if err != nil || len(accounts) <= 0 {
		return nil, err
	}
	return &accounts[0], nil
}
