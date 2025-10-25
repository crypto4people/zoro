package service

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"zoro/service/db"
)

type repo struct {
	mutex    sync.RWMutex
	servers  *db.ServerBox
	subs     *db.SubBox
	txs      *db.TxBox
	settings *db.SettingBox
}

var errServerNotFound = errors.New("server not found")

func (re *repo) UpsertServer(server *db.Server) error {
	if !validateServerURL(server.URL) {
		return errors.New("invalid server url")
	}
	re.mutex.Lock()
	defer re.mutex.Unlock()

	id, err := re.getServer(server.URL)
	if err != nil && !errors.Is(err, errServerNotFound) {
		return err
	}

	err = re.removeServer(server.URL)
	if err != nil {
		return err
	}

	server.ID = id
	server.ID, err = re.servers.Put(server)
	if err != nil {
		return err
	}
	return nil
}

func (re *repo) RemoveServer(url string) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	return re.removeServer(url)
}

func (re *repo) ListServers() ([]db.Server, error) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	return re.servers.GetAll()
}

func (re *repo) ListTxs(cont uint64) ([]db.Tx, error) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	query := re.txs.Query(
		db.Tx_.ID.LessOrEqual(cont),
		db.Tx_.ID.OrderDesc(),
	)
	defer query.Close()
	query.Limit(10)
	return query.Find()
}

func (re *repo) UpdateTx(tx *db.Tx) error {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	return re.txs.Update(tx)
}

func (re *repo) Clean() error {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	panic("todo")
}

func (re *repo) getServer(url string) (uint64, error) {
	query := re.servers.Query(db.Server_.URL.Equals(url, true))
	defer query.Close()

	ids, err := query.FindIds()

	if err != nil {
		return 0, err
	}

	switch len(ids) {
	case 0:
		return 0, errServerNotFound
	case 1:
		return ids[0], nil
	default:
		_, err = re.servers.RemoveIds(ids[1:]...)
		return ids[0], err
	}
}

func (re *repo) removeServer(url string) error {
	ids, err := re.getServerSubs(url)
	if err != nil {
		return err
	}

	id, err := re.getServer(url)
	if err != nil {
		if errors.Is(err, errServerNotFound) {
			return nil
		}
	}

	if len(ids) > 0 {
		_, err = re.subs.RemoveIds(ids...)
		if err != nil {
			return err
		}
	}

	return re.servers.RemoveId(id)
}

func (re *repo) getServerSubs(url string) ([]uint64, error) {
	query := re.subs.Query(db.Sub_.Server.Equals(url, true))
	defer query.Close()
	return query.FindIds()
}

func validateServerURL(str string) bool {
	if str == "" {
		return false
	}
	if strings.ToLower(str) != str {
		return false
	}
	purl, err := url.Parse(str)
	if err != nil {
		return false
	}
	if purl.String() != str {
		return false
	}
	return true
}
