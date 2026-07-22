package handler

import (
	"html/template"
	"net/http"

	"github.com/yusiwen/myops/internal/config"
	"github.com/yusiwen/myops/internal/middleware"
)

type SettingsHandler struct {
	cfg  *config.Config
	tmpl *template.Template
}

func NewSettingsHandler(cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{
		cfg:  cfg,
		tmpl: template.Must(template.ParseFiles("web/templates/base.html", "web/templates/settings.html")),
	}
}

type settingsData struct {
	User       *middleware.SessionUser
	GiteaURL   string
	DroneURL   string
	DroneToken string
}

func (h *SettingsHandler) Page(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	data := settingsData{
		User:       user,
		GiteaURL:   h.cfg.Gitea.URL,
		DroneURL:   h.cfg.Drone.URL,
		DroneToken: maskToken(h.cfg.Drone.Token),
	}

	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.ExecuteTemplate(w, "content", data)
		return
	}
	h.tmpl.Execute(w, data)
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
