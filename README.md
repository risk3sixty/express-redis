# express-redis

Redis session store compatible with [express-session-go](https://github.com/whatl3y/express-session-go)

## Usage

See `examples/server.go`

```go
package main

import (
	"log"
	"net/http"

	s "github.com/risk3sixty/express-redis/store"
	m "github.com/whatl3y/express-session-go/middleware"
)

func handler(w http.ResponseWriter, r *http.Request) {
	session, _ := r.Context().Value(m.SessionContextKey).(m.Session)
	sid := session.SessionID
	// session.SessionData contains all stored session data
	w.Write([]byte("Session ID: " + sid))
}

func main() {
	var redisStore s.RedisStore
	redisStore.CreateClient("redis://localhost:6379")

	m.SetStore(&redisStore)
	m.SetCookieKey("sid")
	m.SetCookieSecret("r3stesting123")

	final := m.ExpressSessionMiddleware(http.HandlerFunc(handler))
	http.Handle("/", final)
	http.ListenAndServe(":8080", nil)
}
```
