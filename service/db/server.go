package db

import (
	"errors"

	"github.com/crypto4people/zoro/service/db/ob"
	"github.com/samber/lo"
)

func (db *Database) ListServers() (map[string]string, error) {
	servers, err := db.servers.GetAll()
	if err != nil {
		return nil, err
	}

	return lo.SliceToMap(servers, func(server ob.Server) (string, string) {
		return server.Name, server.URL
	}), nil
}

func (db *Database) SaveServer(name string, serverURL string) error {
	server, err := db.servers.Query(ob.Server_.Name.Equals(name, false)).Find()
	if err != nil {
		return err
	}

	if len(server) == 0 {
		newServer := &ob.Server{
			Name: name,
			URL:  serverURL,
		}
		_, err = db.servers.Put(newServer)
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	server[0].URL = serverURL
	_, err = db.servers.Put(&server[0])

	return err
}

func (db *Database) DeleteServer(name string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	server, err := db.servers.Query(ob.Server_.Name.Equals(name, false)).Find()
	if err != nil {
		return err
	}

	if len(server) == 0 {
		return nil
	}

	return db.servers.Remove(&server[0])
}

func (db *Database) GetServer(name string) (string, error) {
	server, err := db.servers.Query(ob.Server_.Name.Equals(name, false)).Find()
	if err != nil {
		return "", err
	}

	if len(server) == 0 {
		return "", errors.New("not found")
	}

	return server[0].URL, nil
}
