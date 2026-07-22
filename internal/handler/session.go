package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store sessions.Store

func InitSessionStore(key string) sessions.Store {
	if len(key) < 32 {
		log.Fatalf("SESSION_KEY must be at least 32 characters, got %d", len(key))
	}
	cs := sessions.NewCookieStore([]byte(key))
	cs.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	store = cs
	return store
}

func GetStore() sessions.Store {
	return store
}
