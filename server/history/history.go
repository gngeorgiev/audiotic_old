package history

import (
	"gngeorgiev/audiotic/server/models"
	"sync"

	"github.com/asdine/storm"
)

var (
	once sync.Once
	db   *storm.DB
)

func Init() error {
	var err error
	once.Do(func() {
		d, e := storm.Open("db.db")
		if e != nil {
			err = e
			return
		}

		db = d
	})

	if db != nil {
		if err := db.Init(models.Track{}); err != nil {
			return err
		}
	}

	return err
}

func Release() error {
	return db.Close()
}

func Add(t *models.Track) error {
	var existing interface{}
	if err := db.Find("id", t.ID, existing); err == nil && existing != nil {
		if updateErr := db.UpdateField(t, "LastPlayed", t.LastPlayed); updateErr != nil {
			return updateErr
		}
	}

	return db.Save(t)
}

func Get() ([]models.Track, error) {
	var res []models.Track
	if err := db.AllByIndex("LastPlayed", &res, storm.Reverse()); err != nil {
		return nil, err
	}

	return res, nil
}
