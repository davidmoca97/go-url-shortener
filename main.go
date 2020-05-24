package main

import (
	"os"
	"context"
	"log"
	"net/url"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	var err error
	DBClient, err = InitializeDB()
	if err != nil {
		log.Fatal("Could not connect to the database")
	}
	defer DBClient.Disconnect(context.Background())
	
	router := mux.NewRouter()
	staticDir := "/static/"
	
	router.
	PathPrefix(staticDir).
	Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("."+staticDir))))
	router.HandleFunc("/url-shortener", urlShortener).Methods(http.MethodPost)
	router.HandleFunc("/{id}", redirectToOriginalURL).Methods(http.MethodGet)
	router.HandleFunc("/", serveIndexPage).Methods(http.MethodGet)
	
	http.ListenAndServe(":9999", router)
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

	//Check if the url already exists in the db
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
		hostName, _ := os.LookupEnv("HOST_URI")
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
	if err == mongo.ErrNoDocuments {
		log.Printf("Hash not found: %s", id)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
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
	salt, _ := os.LookupEnv("HASH_SALT")
	hd.Salt = salt
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
