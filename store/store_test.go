package store

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

func TestGetWithNoSession(t *testing.T) {
	store := getStore(t)
	shouldBeNil, err := store.Get("abc123")
	if err != nil {
		t.Error("error getting empty session", err)
	}

	if shouldBeNil != nil {
		t.Error("session is not nil", shouldBeNil)
	}
}

func TestSetGetWithTinySession(t *testing.T) {
	store := getStore(t)
	populateTinySession("abc123", t, &store)

	session, err := store.Get("abc123")
	if err != nil {
		t.Error("error getting valid session", err)
	}

	shouldBeMap := session.(map[string]interface{})
	if shouldBeMap["r3s"] != "team" {
		t.Error("session key is not correct", shouldBeMap)
	}
}

func TestDestroy(t *testing.T) {
	sid := "newsession"
	store := getStore(t)

	err := store.Destroy(sid)
	if err != nil {
		t.Error("error destroying unknown session", err)
	}

	populateTinySession(sid, t, &store)

	err = store.Destroy(sid)
	if err != nil {
		t.Error("error destroying session", err)
	}
}

func TestAll(t *testing.T) {
	store := getStore(t)
	noKeys, err := store.All()
	if err != nil {
		t.Error("Error getting no keys", nil)
	}

	if noKeys != nil {
		t.Error("should be no keys in the DB", noKeys)
	}

	populateTinySession("abc123", t, &store)
	singleSessionSlice, err := store.All()
	if err != nil {
		t.Error("Error getting single key", nil)
	}

	if singleSessionSlice[0] != "sess:abc123" {
		t.Error("incorrect value for single session key", singleSessionSlice)
	}
}

func TestClear(t *testing.T) {
	store := getStore(t)
	populateTinySession("newsession", t, &store)

	keys, err := store.All()
	if err != nil {
		t.Error("error getting sessions", err)
	}
	if len(keys) != 1 {
		t.Error("number of keys is not correct", keys)
	}

	err = store.Clear()
	if err != nil {
		t.Error("clearing sessions returned error", err)
	}

	keys, err = store.All()
	if err != nil {
		t.Error("clearing sessions returned error", err)
	}
	if len(keys) != 0 {
		t.Error("number of keys after clear is not correct", keys)
	}
}

func TestLength(t *testing.T) {
	store := getStore(t)

	numKeys, err := store.Length()
	if err != nil {
		t.Error("error getting sessions", err)
	}
	if numKeys != 0 {
		t.Error("number of keys is not correct", numKeys)
	}

	populateTinySession("newsession", t, &store)
	numKeys, err = store.Length()
	if err != nil {
		t.Error("error getting sessions", err)
	}
	if numKeys != 1 {
		t.Error("number of keys is not correct", numKeys)
	}
}

// func TestTouch(t *testing.T) {

// }

func TestGetTTL(t *testing.T) {
	sid := "abc123"
	store := getStore(t)

	ttl, err := store.getTTL(sid)
	if err != nil {
		t.Error("no cookie returns error", err)
	}

	if ttl != 0 {
		t.Error("no cookie returns nonzero ttl", ttl)
	}

	populateTinySession(sid, t, &store)
	ttl, err = store.getTTL(sid)
	if err != nil {
		t.Error("no cookie returns error", err)
	}

	if ttl != 0 {
		t.Error("no cookie returns nonzero ttl", ttl)
	}

	populateTinySessionWithCookie(sid, t, &store)
	ttl, err = store.getTTL(sid)
	if err != nil {
		t.Error("valid cookie returns error", err)
	}

	if ttl < 599 { // should be 10 minutes in seconds (i.e. 600)
		t.Error("valid cookie returns non positive ttl", ttl)
	}
}

func newTestRedis() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client
}

func getStore(t *testing.T) RedisStore {
	var store RedisStore
	client := newTestRedis()
	err := store.SetClient(client)
	if err != nil {
		t.Error("Error getting no keys", nil)
	}
	return store
}

func populateTinySession(sid string, t *testing.T, store *RedisStore) {
	myMap := make(map[string]interface{})
	myMap["r3s"] = "team"
	err := store.Set(sid, myMap)
	if err != nil {
		t.Error("error setting session", err)
	}
}

func populateTinySessionWithCookie(sid string, t *testing.T, store *RedisStore) {
	cookie := make(map[string]interface{})
	cookie["expires"] = time.Now().Add(10 * time.Minute)

	myMap := make(map[string]interface{})
	myMap["r3s"] = "team"
	myMap["cookie"] = cookie
	err := store.Set(sid, myMap)
	if err != nil {
		t.Error("error setting session", err)
	}
}
