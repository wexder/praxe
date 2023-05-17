package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type URL struct {
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

var urlMap = make(map[string]string)

func generateShortURL() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURL := make([]byte, 6)
	for i := range shortURL {
		shortURL[i] = letters[rand.Intn(len(letters))]
	}
	return string(shortURL)
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var url URL
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if url.LongURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "LongURL cannot be empty")
		return
	}

	shortURL := generateShortURL()
	urlMap[shortURL] = url.LongURL

	response := URL{LongURL: url.LongURL, ShortURL: shortURL}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	longURL, ok := urlMap[shortURL]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "URL not found")
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	rand.Seed(42) // For consistent short URL generation in this example

	http.HandleFunc("/shorten", shortenURLHandler)
	http.HandleFunc("/", redirectHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
