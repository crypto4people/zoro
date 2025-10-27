package db

import (
	"net/url"
	"zoro/service/db/ob"
)

func (db *Database) Servers() (map[string]*url.URL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	servers, err := db.servers.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[string]*url.URL, len(servers))
	for _, server := range servers {
		purl, err := url.Parse(server.URL)
		if err != nil {
			return nil, err
		}
		result[server.Name] = purl
	}
	return result, nil
}

func (db *Database) DeleteServer(name string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	server, err := db.getServer(name)
	if err == nil && server != nil {
		err = db.servers.RemoveId(server.ID)
	}
	return err
}

func (db *Database) PutServer(name string, serverURL *url.URL) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	server, err := db.getServer(name)
	if err != nil {
		return err
	}
	if server == nil {
		server = &ob.Server{
			Name: name,
			URL:  serverURL.String(),
		}
	} else {
		server.URL = serverURL.String()
	}
	_, err = db.servers.Put(server)
	return err
}

func (db *Database) GetServer(name string) (*url.URL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	server, err := db.getServer(name)
	if err != nil || server == nil {
		return nil, err
	}
	return url.Parse(server.URL)
}

func (db *Database) getServer(key string) (*ob.Server, error) {
	query := db.servers.Query(ob.Server_.Name.Equals(key, true))
	defer query.Close()
	servers, err := query.Limit(1).Find()
	if err != nil || len(servers) <= 0 {
		return nil, err
	}
	return &servers[0], nil
}
