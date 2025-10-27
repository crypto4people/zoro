package db

import (
	"encoding/hex"
	"net/url"
	"zoro/service/db/ob"

	"github.com/ethereum/go-ethereum/common"
)

func (db *Database) Pends() (map[common.Hash]*url.URL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	pends, err := db.pends.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]*url.URL, len(pends))
	for _, pend := range pends {
		purl, err := url.Parse(pend.ServerURL)
		if err != nil {
			return nil, err
		}
		result[common.HexToHash(pend.Hash)] = purl
	}
	return result, nil
}

func (db *Database) DeletePend(hash common.Hash) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	pend, err := db.getPend(hex.EncodeToString(hash[:]))
	if err == nil && pend != nil {
		err = db.pends.RemoveId(pend.ID)
	}
	return err
}

func (db *Database) PutPend(hash common.Hash, serverURL *url.URL) error {

	hashHex := hex.EncodeToString(hash[:])

	db.mutex.Lock()
	defer db.mutex.Unlock()

	pend, err := db.getPend(hashHex)
	if err != nil {
		return err
	}
	if pend == nil {
		pend = &ob.Pend{
			Hash:      hashHex,
			ServerURL: serverURL.String(),
		}
	} else {
		pend.ServerURL = serverURL.String()
	}

	_, err = db.pends.Put(pend)
	return err
}

func (db *Database) GetPend(hash common.Hash) (*url.URL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	pend, err := db.getPend(hex.EncodeToString(hash[:]))
	if err != nil {
		return nil, err
	}
	return url.Parse(pend.ServerURL)
}

func (db *Database) getPend(hash string) (*ob.Pend, error) {
	query := db.pends.Query(ob.Pend_.Hash.Equals(hash, true))
	defer query.Close()
	pends, err := query.Limit(1).Find()
	if err != nil || len(pends) <= 0 {
		return nil, err
	}
	return &pends[0], nil
}
