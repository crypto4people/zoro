package db

import (
	"errors"
	"zoro/service/db/ob"
)

func (db *Database) Servers() (map[string]string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	servers, err := db.servers.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(servers))
	for _, server := range servers {
		result[server.Key] = server.URL
	}
	return result, nil
}

func (db *Database) DeleteServer(key string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	server, err := db.getServer(key)
	if err == nil && server != nil {
		err = db.servers.RemoveId(server.ID)
	}
	return err
}

func (db *Database) PutServer(key string, url string) error {
	if !validKey(key) {
		return errors.New("invalid key")
	}

	if !validURL(url) {
		return errors.New("invalid url")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	server, err := db.getServer(key)
	if err != nil {
		return err
	}
	if server == nil {
		server = &ob.Server{
			Key: key,
			URL: url,
		}
	} else {
		server.URL = url
	}
	_, err = db.servers.Put(server)
	return err
}

func (db *Database) GetServer(key string) (string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	server, err := db.getServer(key)
	if err != nil {
		return "", err
	}
	return server.URL, nil
}

func (db *Database) getServer(key string) (*ob.Server, error) {
	query := db.servers.Query(ob.Server_.Key.Equals(key, true))
	defer query.Close()
	servers, err := query.Limit(1).Find()
	if err != nil || len(servers) <= 0 {
		return nil, err
	}
	return &servers[0], nil
}
