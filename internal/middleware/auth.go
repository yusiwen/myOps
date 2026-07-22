package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type SessionUser struct {
	Login     string
	AvatarURL string
	Token     string
}

type contextKey string

const userKey contextKey = "user"

func Auth(sessionName string, store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, _ := store.Get(r, sessionName)
			user, ok := sess.Values["user"].(*SessionUser)
			if !ok || user == nil {
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", "/")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetUser(r *http.Request, user *SessionUser) *http.Request {
	ctx := context.WithValue(r.Context(), userKey, user)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *SessionUser {
	user, _ := r.Context().Value(userKey).(*SessionUser)
	return user
}
