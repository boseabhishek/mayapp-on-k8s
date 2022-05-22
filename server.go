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

var db *DB

//Configuration is a struct for holding the application configuration data read from a JSON file
type Configuration struct {
	RedisHost string
	RedisPort string
}

func init() {
	// filename is the path to the json config file
	// this file can either be found on local or copied inside container or as configmap inside k8s cluster
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("error opening config.json file %v", err)
	}
	decoder := json.NewDecoder(file)
	var configuration Configuration
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatalf("error decoding config.json file %v", err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" {
		if redisHost = configuration.RedisHost; redisHost == "" {
			redisHost = "localhost"
		}
	}
	if redisPort == "" {
		if redisPort = configuration.RedisPort; redisPort == "" {
			redisPort = "6379"
		}
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0, // use default DB
	})
	db = &DB{}
	db.redis = rdb
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/cats/{id}", getCats(db))
	router.HandleFunc("/cats", postCats(db))

	fmt.Println("starting cats server...✅✅✅✅✅")
	log.Fatal(http.ListenAndServe(":8080", router))
}

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
	fmt.Println("entering getCats handler... ✨✨✨✨✨✨✨")
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
	fmt.Println("entering postCats handler... ✨✨✨✨✨✨✨")
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
