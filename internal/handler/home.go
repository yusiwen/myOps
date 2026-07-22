package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/yusiwen/myops/internal/drone"
	"github.com/yusiwen/myops/internal/gitea"
	applog "github.com/yusiwen/myops/internal/log"
	"github.com/yusiwen/myops/internal/middleware"
)

type HomeHandler struct {
	gitea *gitea.Client
	drone *drone.Client
	log   *applog.Logger
	tmpl  *template.Template
}

func NewHomeHandler(gitea *gitea.Client, drone *drone.Client, logger *applog.Logger) *HomeHandler {
	return &HomeHandler{
		gitea: gitea,
		drone: drone,
		log:   logger,
		tmpl:  template.Must(template.ParseFiles("web/templates/base.html", "web/templates/dashboard.html")),
	}
}

type buildRow struct {
	Owner   string
	Repo    string
	Number  int64
	Event   string
	Status  string
	Branch  string
	Author  string
	Created string
	created int64
}

type dashboardData struct {
	User       *middleware.SessionUser
	Skeleton   bool
	RepoCount  int
	BuildCount int
	Builds     []buildRow
	RepoError  string
	DroneError string
}

func (h *HomeHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	partial := r.URL.Query().Get("partial") == "1"

	if !partial {
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(skeletonHTML("/dashboard")))
			return
		}
		h.tmpl.Execute(w, map[string]interface{}{
			"User": user, "Loading": true, "SelfURL": "/dashboard",
		})
		return
	}

	data := dashboardData{User: user}

	if h.gitea != nil {
		n, err := h.gitea.RepoCount()
		if err != nil {
			data.RepoError = err.Error()
			h.log.Error("[dashboard] RepoCount: %v", err)
		} else {
			data.RepoCount = n
			h.log.Info("[dashboard] repos=%d", n)
		}
	}

	if h.drone != nil {
		repos, err := h.drone.ListRepos()
		if err != nil {
			data.DroneError = err.Error()
			h.log.Error("[dashboard] RepoList: %v", err)
		} else {
			var totalBuilds int64
			var builds []buildRow
			var needBuildLast []string

			for _, repo := range repos {
				totalBuilds += repo.Counter
				if repo.Build.Number > 0 {
					builds = append(builds, buildRow{
						Owner:   repo.Namespace,
						Repo:    repo.Name,
						Number:  repo.Build.Number,
						Event:   repo.Build.Event,
						Status:  repo.Build.Status,
						Branch:  repo.Build.Target,
						Author:  repo.Build.Author,
						created: repo.Build.Created,
					})
				} else if repo.Counter > 0 {
					needBuildLast = append(needBuildLast, repo.Namespace+"/"+repo.Name)
				}
			}

			if len(needBuildLast) > 0 {
				h.log.Info("[dashboard] fetching BuildLast for %d repos", len(needBuildLast))
				for _, ns := range needBuildLast {
					parts := splitRepo(ns)
					if len(parts) < 2 {
						continue
					}
					b, err := h.drone.BuildLast(parts[0], parts[1])
					if err != nil {
						continue
					}
					if b != nil && b.Number > 0 {
						builds = append(builds, buildRow{
							Owner:   parts[0],
							Repo:    parts[1],
							Number:  b.Number,
							Event:   b.Event,
							Status:  b.Status,
							Branch:  b.Target,
							Author:  b.Author,
							created: b.Created,
						})
					}
				}
			}

			sort.Slice(builds, func(i, j int) bool {
				return builds[i].created > builds[j].created
			})
			for i := range builds {
				builds[i].Created = timeFmt(builds[i].created)
			}
			if len(builds) > 10 {
				builds = builds[:10]
			}
			data.Builds = builds
			data.BuildCount = int(totalBuilds)
			h.log.Info("[dashboard] repos=%d builds=%d recent=%d",
				data.RepoCount, data.BuildCount, len(data.Builds))
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.ExecuteTemplate(w, "content", data)
		return
	}
	h.tmpl.Execute(w, data)
}

func splitRepo(s string) []string {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		return nil
	}
	return []string{s[:i], s[i+1:]}
}

func skeletonHTML(selfURL string) string {
	return `<div class="flex items-center justify-center py-20 text-gray-400 dark:text-gray-500">
  <svg class="animate-spin h-8 w-8 mr-3" viewBox="0 0 24 24" fill="none">
    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
  </svg>
  <span>Loading...</span>
</div>
<div hx-get="` + selfURL + `?partial=1" hx-trigger="load delay:50ms" hx-target="#main" hx-swap="innerHTML" hx-push-url="false"></div>`
}

func timeFmt(ts int64) string {
	t := time.Unix(ts, 0)
	now := time.Now()
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		return t.Format("Jan 2")
	}
}
