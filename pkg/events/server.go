package events

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	Database *gorm.DB = nil
)

func HandlePublish(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	key := r.FormValue("key")

	var access LiveAccess
	Database.Where(&LiveAccess{Key: key}).First(&access)
	if len(access.Key) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	access.StreamName = name
	access.Online = true
	Database.Save(&access)

	w.WriteHeader(http.StatusOK)
}

func HandlePlay(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	var access LiveAccess
	Database.Where(&LiveAccess{StreamName: name}).First(&access)
	if len(access.StreamName) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	access.Viewers++
	Database.Save(&access)

	w.WriteHeader(http.StatusOK)
}

func HandlePlayDone(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	var access LiveAccess
	Database.Where(&LiveAccess{StreamName: name}).First(&access)
	if len(access.StreamName) <= 0 || !access.Online {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	access.Viewers--
	Database.Save(&access)

	w.WriteHeader(http.StatusOK)
}

func HandlePublishDone(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")

	var access LiveAccess
	Database.Where(&LiveAccess{Key: key}).First(&access)
	if len(access.Key) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	access.Online = false
	access.Viewers = 0
	Database.Save(&access)

	w.WriteHeader(http.StatusOK)
}

func Init(db *sql.DB) error {
	ormDb, err := gorm.Open("mysql", db)
	if err != nil {
		return err
	}

	Database = &ormDb
	return nil
}
