package store

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

func TestGetWithNoSession(t *testing.T) {
	var store RedisStore
	client := newTestRedis()
	err := store.SetClient(client)
	if err != nil {
		t.Error("error setting store client", err)
	}

	shouldBeNil, err := store.Get("abc123")
	if err != nil {
		t.Error("error getting empty session", err)
	}

	if shouldBeNil != nil {
		t.Error("session is not nil", shouldBeNil)
	}
}

func TestSetGetWithTinySession(t *testing.T) {
	var store RedisStore
	client := newTestRedis()
	err := store.SetClient(client)
	if err != nil {
		t.Error("error setting store client", err)
	}

	myMap := make(map[string]interface{})
	myMap["r3s"] = "team"
	err = store.Set("abc123", myMap)
	if err != nil {
		t.Error("error setting session", err)
	}

	session, err := store.Get("abc123")
	if err != nil {
		t.Error("error getting valid session", err)
	}

	shouldBeMap := session.(map[string]interface{})
	if shouldBeMap["r3s"] != "team" {
		t.Error("session is not nil", shouldBeMap)
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
