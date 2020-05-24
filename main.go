package main

import (
	"net/url"
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
	server.HandleFunc("/", serveIndexPage).Methods(http.MethodGet)
	
	staticDir := "/static/"
	server.
		PathPrefix(staticDir).
		Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("."+staticDir))))
	
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
		http.Error(w, "There was a problem decoding the body of the request", http.StatusInternalServerError)
		return
	}
	if !validateURL(record.OriginalURL) {
		http.Error(w, "The provided URL is not valid", http.StatusBadRequest)
		return
	}
	h, err := getHashID()
	if err != nil {
		http.Error(w, "Error getting the hash", http.StatusInternalServerError)
		return
	}

	currentTime := time.Now().Nanosecond()
	hash, err := h.Encode([]int{int(currentTime)})
	if err != nil {
		http.Error(w, "Error getting the hash", http.StatusInternalServerError)
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

func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
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

func validateURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}