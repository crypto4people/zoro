package db

import (
	"github.com/crypto4people/zoro/service/db/ob"
)

func (db *Database) Settings() (map[string]string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	pends, err := db.settings.GetAll()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(pends))
	for _, setting := range pends {
		result[setting.Key] = setting.Value
	}
	return result, nil
}

func (db *Database) PutSetting(key string, val string) error {

	if len(val) <= 0 {
		return db.deleteSetting(key)
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	setting, err := db.getSetting(key)
	if err != nil {
		return err
	}
	if setting == nil {
		setting = &ob.Setting{
			Key:   key,
			Value: val,
		}
	} else {
		setting.Value = val
	}
	_, err = db.settings.Put(setting)
	return err
}

func (db *Database) deleteSetting(hash string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	setting, err := db.getSetting(hash)
	if err == nil && setting != nil {
		err = db.settings.RemoveId(setting.ID)
	}
	return err
}

func (db *Database) getSetting(key string) (*ob.Setting, error) {
	query := db.settings.Query(ob.Setting_.Key.Equals(key, true))
	defer query.Close()
	settings, err := query.Limit(1).Find()
	if err != nil || len(settings) <= 0 {
		return nil, err
	}
	return &settings[0], nil
}
