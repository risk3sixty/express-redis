package store

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/go-redis/redis"
)

var client *redis.Client

// RedisStore is the memory store
// NOTE: really shouldn't be used in prod
type RedisStore map[string]interface{}

// SetClient allows us to use a client that the user already
// has that resolves to the redis.Client interface
func (store *RedisStore) SetClient(redisClient interface{}) error {
	var ok bool
	client, ok = redisClient.(*redis.Client)
	if !ok {
		return errors.New("client not valid")
	}
	return nil
}

// CreateClient initializes the redis client the store will use
func (store *RedisStore) CreateClient(connString string) error {
	connStruct, err := url.Parse(connString)
	if err != nil {
		return err
	}

	db := 0
	if connStruct.Path != "" {
		db, err = strconv.Atoi(connStruct.Path)
		if err != nil {
			return err
		}
	}

	password, _ := connStruct.User.Password()

	client = redis.NewClient(&redis.Options{
		Addr:     connStruct.Host,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	return nil
}

// Destroy removes a session from the store
func (store *RedisStore) Destroy(sid string) error {
	_, err := client.Del("sess:" + sid).Result()
	return err
}

// Get retrieves the session from the store
func (store *RedisStore) Get(sid string) (interface{}, error) {
	jsonSession, err := client.Get("sess:" + sid).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if jsonSession == "" {
		return nil, nil
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonSession), &jsonMap)
	return jsonMap, err
}

// Set sets a session with a given value
func (store *RedisStore) Set(sid string, jsonMap map[string]interface{}) error {
	jsonSession, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}

	_, err = client.Set("sess:"+sid, string(jsonSession), 0).Result()
	return err
}
