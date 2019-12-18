package store

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

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
		Password: password, // "" is no password set
		DB:       db,       // 0 is default DB
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

// All gets all sessions in the DB
func (store *RedisStore) All() ([]string, error) {
	var keys []string
	return scanKeys("sess:*", keys, uint64(0))
}

// Clear gets all session IDs and deletes them
// from the store
func (store *RedisStore) Clear() error {
	keys, err := store.All()
	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err = client.Del(key).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

// Length gets the number of sessions in the store
func (store *RedisStore) Length() (int, error) {
	keys, err := store.All()
	if err != nil {
		return 0, nil
	}
	return len(keys), nil
}

// Touch updates the redis TTL for a session when
// the session hits the server
func (store *RedisStore) Touch(sid string) error {
	ttl, err := store.getTTL(sid)
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration(string(ttl) + "s")
	if err != nil {
		return err
	}

	_, err = client.Expire("sess:"+sid, dur).Result()
	return err
}

func (store *RedisStore) getTTL(sid string) (int64, error) {
	session, err := store.Get(sid)
	if err != nil {
		return 0, err
	}

	sessionMap, ok := session.(map[string]interface{})
	if !ok {
		return 0, nil
	}

	cookie := sessionMap["cookie"]
	if cookie == nil {
		return 0, nil
	}

	cookieMap := cookie.(map[string]interface{})
	expires := cookieMap["expires"].(string)
	expiresTs, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		return 0, err
	}

	unixExpires := expiresTs.Unix()
	unixNow := time.Now().Unix()
	return unixExpires - unixNow, nil
}

func scanKeys(pattern string, keys []string, cursor uint64) ([]string, error) {
	newKeys, newCursor, err := client.Scan(cursor, pattern, 100).Result()
	if err != nil {
		return keys, err
	}
	combinedKeys := append(keys, newKeys...)
	if newCursor == 0 {
		return combinedKeys, nil
	}

	return scanKeys(pattern, combinedKeys, newCursor)
}
