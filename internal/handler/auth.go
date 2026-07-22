package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	gitea_sdk "code.gitea.io/sdk/gitea"

	"github.com/yusiwen/myops/internal/middleware"
)

func init() {
	gob.Register(&middleware.SessionUser{})
}

type AuthHandler struct {
	giteaURL    string
	oauth2Cfg   *oauth2.Config
	sessionName string
}

func NewAuthHandler(giteaURL, clientID, clientSecret, baseURL, sessionName string) *AuthHandler {
	redirectURL := baseURL + "/auth/callback"
	return &AuthHandler{
		giteaURL:    giteaURL,
		sessionName: sessionName,
		oauth2Cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  giteaURL + "/login/oauth/authorize",
				TokenURL: giteaURL + "/login/oauth/access_token",
			},
			RedirectURL: redirectURL,
			Scopes:      []string{},
		},
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state := randState()
	sess, _ := store.Get(r, h.sessionName)
	sess.Values["oauth_state"] = state
	if err := sess.Save(r, w); err != nil {
		log.Printf("auth: session save error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	log.Printf("auth: saved oauth_state=%q for session", state)

	url := h.oauth2Cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, h.sessionName)

	gotState := r.URL.Query().Get("state")
	sessState, _ := sess.Values["oauth_state"].(string)
	log.Printf("auth: callback state=%q sessionState=%q", gotState, sessState)

	if gotState != sessState {
		http.Error(w, "State mismatch", http.StatusForbidden)
		return
	}
	delete(sess.Values, "oauth_state")

	code := r.URL.Query().Get("code")
	token, err := h.oauth2Cfg.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	client, err := gitea_sdk.NewClient(h.giteaURL, gitea_sdk.SetToken(token.AccessToken))
	if err != nil {
		http.Error(w, fmt.Sprintf("Gitea client error: %v", err), http.StatusInternalServerError)
		return
	}
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Get user failed: %v", err), http.StatusInternalServerError)
		return
	}

	sess.Values["user"] = &middleware.SessionUser{
		Login:     user.UserName,
		AvatarURL: user.AvatarURL,
		Token:     token.AccessToken,
	}
	if err := sess.Save(r, w); err != nil {
		log.Printf("auth: callback session save error: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, h.sessionName)
	delete(sess.Values, "user")
	sess.Options.MaxAge = -1
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("login").Parse(loginHTML))
	tmpl.Execute(w, map[string]string{
		"GiteaURL": h.giteaURL,
	})
}

func GetSession(r *http.Request, name string) *middleware.SessionUser {
	sess, _ := store.Get(r, name)
	user, _ := sess.Values["user"].(*middleware.SessionUser)
	return user
}

var loginHTML = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>myOps</title>
  <link href="/static/css/tailwind.css" rel="stylesheet">
</head>
<body class="bg-surface-dark text-gray-100 flex items-center justify-center min-h-screen">
  <div class="text-center">
    <h1 class="text-4xl font-bold mb-2">myOps</h1>
    <p class="text-gray-400 mb-8">Unified dashboard for Gitea & Drone</p>
    <a href="/auth/login"
       class="inline-block bg-blue-600 hover:bg-blue-500 px-6 py-3 rounded-lg font-semibold transition">
      Sign in with Gitea
    </a>
  </div>
</body>
</html>`

func randState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
