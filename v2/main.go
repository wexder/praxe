package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
)

type URL struct {
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

func generateShortURL() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURL := make([]byte, 6)
	for i := range shortURL {
		shortURL[i] = letters[rand.Intn(len(letters))]
	}
	return string(shortURL)
}

func shortenURLHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if url.ShortURL != "" {
			shortURL = url.ShortURL
		}
		err = rdb.Set(context.Background(), shortURL, url.LongURL, -1).Err()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
}

func redirectHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		longURL, err := rdb.Get(context.Background(), shortURL).Result()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "URL not found")
			return
		}

		http.Redirect(w, r, longURL, http.StatusFound)
	}
}

func main() {
	rand.Seed(42) // For consistent short URL generation in this example
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6390",
	})
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/shorten", shortenURLHandler(rdb))
	http.HandleFunc("/", redirectHandler(rdb))

	log.Fatal(http.ListenAndServe(":8077", nil))
}
