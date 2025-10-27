package db

import (
	"errors"
	"zoro/service/db/ob"
)

func (db *Database) Pends() (map[string]string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	pends, err := db.pends.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(pends))
	for _, pend := range pends {
		result[pend.Hash] = pend.ServerURL
	}
	return result, nil
}

func (db *Database) DeletePend(hash string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	pend, err := db.getPend(hash)
	if err == nil && pend != nil {
		err = db.pends.RemoveId(pend.ID)
	}
	return err
}

func (db *Database) PutPend(hash string, serverURL string) error {
	if !validHash(hash) {
		return errors.New("invalid hash")
	}

	if !validURL(serverURL) {
		return errors.New("invalid server url")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	pend, err := db.getPend(hash)
	if err != nil {
		return err
	}
	if pend == nil {
		pend = &ob.Pend{
			Hash:      hash,
			ServerURL: serverURL,
		}
	} else {
		pend.ServerURL = serverURL
	}
	_, err = db.pends.Put(pend)
	return err
}

func (db *Database) GetPend(hash string) (string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	pend, err := db.getPend(hash)
	if err != nil {
		return "", err
	}
	return pend.ServerURL, nil
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
