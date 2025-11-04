package db

import (
	"errors"

	"github.com/crypto4people/zoro/service/db/ob"
	"github.com/ethereum/go-ethereum/common"
)

func (db *Database) ListAccounts() ([]common.Address, error) {
	accounts, err := db.accounts.GetAll()
	if err != nil {
		return nil, err
	}

	var result []common.Address
	for _, account := range accounts {
		result = append(result, common.HexToAddress(account.Address))
	}

	return result, nil
}

func (db *Database) HasAccount(address common.Address) (bool, error) {
	account, err := db.accounts.Query(ob.Account_.Address.Equals(address.Hex(), false)).Find()
	if err != nil {
		return false, err
	}

	return len(account) > 0, nil
}

func (db *Database) LoadAccount(address common.Address) ([]byte, error) {
	account, err := db.accounts.Query(ob.Account_.Address.Equals(address.Hex(), false)).Find()
	if err != nil {
		return nil, err
	}

	if len(account) == 0 {
		return nil, errors.New("not found")
	}

	return account[0].Cipher, nil
}

func (db *Database) DeleteAccount(address common.Address) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	account, err := db.accounts.Query(ob.Account_.Address.Equals(address.Hex(), false)).Find()
	if err != nil {
		return err
	}

	if len(account) == 0 {
		return errors.New("not found")
	}

	return db.accounts.Remove(&account[0])
}

func (db *Database) SaveAccount(address common.Address, cipher []byte) error {
	account, err := db.accounts.Query(ob.Account_.Address.Equals(address.Hex(), false)).Find()
	if err != nil {
		return err
	}

	if len(account) == 0 {
		newAccount := &ob.Account{
			Address: address.Hex(),
			Cipher:  cipher,
		}
		_, err = db.accounts.Put(newAccount)
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	account[0].Cipher = cipher
	_, err = db.accounts.Put(&account[0])
	return err
}
