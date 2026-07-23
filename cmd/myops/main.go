package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/yusiwen/myops/internal/code"
	"github.com/yusiwen/myops/internal/config"
	"github.com/yusiwen/myops/internal/drone"
	"github.com/yusiwen/myops/internal/gitea"
	"github.com/yusiwen/myops/internal/handler"
	applog "github.com/yusiwen/myops/internal/log"
	"github.com/yusiwen/myops/internal/middleware"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.StringVar(configPath, "c", "", "path to config file (shorthand)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger, err := applog.New(cfg.LogFile)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Close()

	logger.Info("=== myOps starting ===")
	logger.Info("listen_addr=%s base_url=%s log_file=%s", cfg.ListenAddr, cfg.BaseURL, cfg.LogFile)

	if cfg.SessionKey == "" {
		logger.Error("SESSION_KEY is required")
		log.Fatal("SESSION_KEY is required")
	}
	sessionStore := handler.InitSessionStore(cfg.SessionKey)

	var giteaClient *gitea.Client
	if cfg.Gitea.URL != "" {
		giteaClient = gitea.New(cfg.Gitea.URL, cfg.Gitea.Token, logger)
	}
	if giteaClient == nil {
		logger.Info("[gitea] not configured, Gitea features disabled")
	}

	var droneClient *drone.Client
	if cfg.Drone.URL != "" {
		droneClient = drone.New(cfg.Drone.URL, cfg.Drone.Token, logger)
	}

	if droneClient == nil {
		logger.Info("[drone] not configured, Drone features disabled")
	}

	if err := code.GenerateChromaCSS("web/static/css/chroma.css", "github", "dracula"); err != nil {
		logger.Info("chroma css generation: %v", err)
	}

	var renderer code.CodeRenderer
	renderer = code.NewChromaRenderer("github", "dracula")

	authHandler := handler.NewAuthHandler(cfg.Gitea.URL, cfg.Gitea.ClientID, cfg.Gitea.ClientSecret, cfg.BaseURL, "myops")
	homeHandler := handler.NewHomeHandler(giteaClient, droneClient, logger)
	repoHandler := handler.NewRepoHandler(giteaClient, renderer, logger)
	buildHandler := handler.NewBuildHandler(droneClient, logger)
	settingsHandler := handler.NewSettingsHandler(cfg)

	authMw := middleware.Auth("myops", sessionStore)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.GetStore().Get(r, "myops")
		u, ok := sess.Values["user"].(*middleware.SessionUser)
		if !ok || u == nil {
			authHandler.LoginPage(w, r)
			return
		}
		homeHandler.Dashboard(w, middleware.SetUser(r, u))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
		r.Get("/logout", authHandler.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Get("/dashboard", homeHandler.Dashboard)
		r.Get("/repos", repoHandler.List)
		r.Get("/repos/{owner}/{name}", repoHandler.Detail)
		r.Get("/repos/{owner}/{name}/file", repoHandler.File)
		r.Get("/builds", buildHandler.List)
		r.Get("/builds/{owner}/{name}/{number}", buildHandler.Detail)
		r.Get("/settings", settingsHandler.Page)
	})

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	logger.Out("myOps listening on %s", cfg.ListenAddr)
	logger.Info("listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, r); err != nil {
		logger.Error("server error: %v", err)
		log.Fatal(err)
	}
}
