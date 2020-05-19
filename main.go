package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
)

const hostName = "http://localhost:9999"

type DataBase struct {
	database []Record
}

func (db *DataBase) Insert(r Record) []Record {
	db.database = append(db.database, r)
	return db.database
}

func (db *DataBase) GetAll() []Record {
	return db.database
}

func (db *DataBase) GetByID(ID string) (Record, int) {
	for idx, r := range db.database {
		if r.ID == ID {
			return r, idx
		}
	}
	return Record{}, -1
}

var DB DataBase

func main() {
	server := mux.NewRouter()
	server.HandleFunc("/url-shortener", urlShortener).Methods(http.MethodPost)
	server.HandleFunc("/{id}", redirectToOriginalURL).Methods(http.MethodGet)
	http.ListenAndServe(":9999", server)
}

type Record struct {
	ID          string `json:"id"`
	OriginalURL string `json:"originalUrl"`
	ShortURL    string `json:"shortUrl"`
}

func urlShortener(w http.ResponseWriter, r *http.Request) {
	var record Record

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&record)
	if err != nil {
		return
	}
	h, err := getHashID()
	if err != nil {
		return
	}

	currentTime := time.Now().Unix()
	hash, err := h.Encode([]int{int(currentTime)})
	if err != nil {
		return
	}

	record.ID = hash
	record.ShortURL = fmt.Sprintf("%s/%s", hostName, hash)
	DB.Insert(record)

	json.NewEncoder(w).Encode(&record)

}

func redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	record, idx := DB.GetByID(id)
	if idx == -1 {
		record.ShortURL = hostName
	}

	http.Redirect(w, r, record.OriginalURL, http.StatusPermanentRedirect)
}

func getHashID() (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = "ilomilo"
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}
	return h, nil
}
