package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type cat struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	HouseName int    `json:"housename"`
}

// DB ...
// TODO: move elsewhere
type DB struct {
	redis *redis.Client
	lock  sync.RWMutex
}

// TODO: different handler file
func getCats(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := mux.Vars(r)
		id := params["id"]

		c, err := fetch(ctx, id, db)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "reason: %v", err)
			return
		}
		// cat transformed to json and encoded to http.ResponseWriter
		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "reason: %v", err)
			return
		}
		return
	}
}

func postCats(db *DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var c cat
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "reason: %v", err)
			return
		}
		id, err := save(ctx, c, db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "reason: %v", err)
			return
		}
		fmt.Fprintf(w, "yayy! cat is saved %s", id)
		return
	}
}

var db *DB

func init() {
	pwd := os.Getenv("REDIS_PASSWORD")
	if pwd == "" {
		fmt.Println("no REDIS_PASSWORD found")
		os.Exit(1)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: pwd,
		DB:       0, // use default DB
	})
	db = &DB{}
	db.redis = rdb
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/cats/{id}", getCats(db))
	router.HandleFunc("/cats", postCats(db))

	fmt.Println("starting cats server...⚡️⚡️⚡️⚡️")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// TODO: different repo file
func save(ctx context.Context, cat cat, db *DB) (string, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, err := json.Marshal(&cat)
	if err != nil {
		return "", fmt.Errorf("error marshalling cat to be saved: %v", cat)
	}

	id := rand.Intn(100) // TODO: fix this to uuid

	err = db.redis.Set(ctx, fmt.Sprintf("%d", id), val, 0).Err()
	if err != nil {
		return "", fmt.Errorf("error saving cat to db: %v", err)
	}
	return fmt.Sprintf("%d", id), nil
}

// TODO: different repo file
func fetch(ctx context.Context, id string, db *DB) (cat, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	res, err := db.redis.Get(ctx, id).Result()
	if err != nil {
		return cat{}, fmt.Errorf("error: %v ; fetching cat by id: %s", err, id)
	}

	var c cat
	err = json.Unmarshal([]byte(res), &c)
	if err != nil {
		return cat{}, fmt.Errorf("error unmarshalling redis data for id: %s", id)
	}

	return c, nil
}
