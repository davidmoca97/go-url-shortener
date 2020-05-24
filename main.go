package main

import (
	"go.mongodb.org/mongo-driver/mongo"
	"context"
	"log"
	"net/url"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
)

const hostName = "http://localhost:9999"

func main() {
	var err error
	DBClient, err = InitializeDB()
	if err != nil {
		log.Fatal("Could not connect to the database")
	}
	defer DBClient.Disconnect(context.Background())
	
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

	res, err := FindByOriginalURL(DBClient, record.OriginalURL)
	if err != nil && err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.ID == "" {
		h, err := getHashID()
		if err != nil {
			http.Error(w, "Error getting the hash", http.StatusInternalServerError)
			return
		}
	
		currentTime := time.Now().Nanosecond()
		hash, err := h.Encode([]int{int(currentTime)})
		if err != nil {
			http.Error(w, "Error getting the", http.StatusInternalServerError)
			return
		}
	
		record.ID = hash
		record.ShortURL = fmt.Sprintf("%s/%s", hostName, hash)
		_, err = Insert(DBClient, record)
		if err != nil {
			http.Error(w, "Database Time out", http.StatusInternalServerError)
			return
		}
	} else {
		record.ID = res.ID
		record.ShortURL = res.ShortURL
	}

	json.NewEncoder(w).Encode(&record)
}

func redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	record, err := GetByID(DBClient, id)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
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