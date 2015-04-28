package events

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	Database *gorm.DB = nil
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {

}

func Listen(db *sql.DB, addr string) error {
	ormDb, err := gorm.Open("mysql", db)
	if err != nil {
		return err
	}
	Database = &ormDb
	defer Database.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", HandleRoot)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(addr, nil))

	return nil
}
