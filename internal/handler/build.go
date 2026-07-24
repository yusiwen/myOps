package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yusiwen/myops/internal/drone"
	applog "github.com/yusiwen/myops/internal/log"
	"github.com/yusiwen/myops/internal/middleware"
)

const buildsPerPage = 50

type BuildHandler struct {
	drone      *drone.Client
	log        *applog.Logger
	listTmpl   *template.Template
	detailTmpl *template.Template
}

func NewBuildHandler(drone *drone.Client, logger *applog.Logger) *BuildHandler {
	return &BuildHandler{
		drone:      drone,
		log:        logger,
		listTmpl:   template.Must(template.ParseFiles("web/templates/base.html", "web/templates/build_list.html")),
		detailTmpl: template.Must(template.ParseFiles("web/templates/base.html", "web/templates/build_detail.html")),
	}
}

type buildItem struct {
	Number int64
	Owner  string
	Repo   string
	Event  string
	Status string
	Branch string
	Author string
	When   string
}

type buildListData struct {
	User       *middleware.SessionUser
	Skeleton   bool
	Builds     []buildItem
	DroneError string
	Page       int
	PrevPage   int
	NextPage   int
	TotalPages int
	TotalCount int
	SelfURL    string
}

func (h *BuildHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	partial := r.URL.Query().Get("partial") == "1"

	if !partial {
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(skeletonHTML("/builds")))
			return
		}
		h.listTmpl.Execute(w, map[string]interface{}{
			"User": user, "Loading": true, "SelfURL": "/builds",
		})
		return
	}

	d := buildListData{User: user}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	d.Page = page

	if page > 1 {
		d.PrevPage = page - 1
	}
	d.NextPage = page + 1

	if h.drone != nil {
		repos, err := h.drone.ListRepos()
		if err != nil {
			h.log.Error("[builds] RepoList: %v", err)
			d.DroneError = err.Error()
		} else {
			type buildRow struct {
				Owner   string
				Repo    string
				Number  int64
				Event   string
				Status  string
				Branch  string
				Author  string
				created int64
			}
			var all []buildRow
			var mu sync.Mutex
			var wg sync.WaitGroup

			for _, repo := range repos {
				if repo.Counter > 0 {
					wg.Add(1)
					go func(owner, repo string) {
						defer wg.Done()
						builds, err := h.drone.ListBuilds(owner, repo)
						if err != nil {
							return
						}
						mu.Lock()
						for _, b := range builds {
							all = append(all, buildRow{
								Owner:   owner,
								Repo:    repo,
								Number:  b.Number,
								Event:   b.Event,
								Status:  b.Status,
								Branch:  b.Target,
								Author:  b.Author,
								created: b.Created,
							})
						}
						mu.Unlock()
					}(repo.Namespace, repo.Name)
				}
			}
			wg.Wait()

			sort.Slice(all, func(i, j int) bool {
				return all[i].created > all[j].created
			})

			d.TotalCount = len(all)
			d.TotalPages = (d.TotalCount + buildsPerPage - 1) / buildsPerPage

			if d.TotalPages > 0 && page > d.TotalPages {
				page = d.TotalPages
				d.Page = page
			}

			start := (page - 1) * buildsPerPage
			end := start + buildsPerPage
			if end > d.TotalCount {
				end = d.TotalCount
			}

			if start < d.TotalCount {
				for _, b := range all[start:end] {
					d.Builds = append(d.Builds, buildItem{
						Number: b.Number,
						Owner:  b.Owner,
						Repo:   b.Repo,
						Event:  b.Event,
						Status: b.Status,
						Branch: b.Branch,
						Author: b.Author,
						When:   timeFmt(b.created),
					})
				}
			}

			h.log.Info("[builds] total=%d page=%d/%d items=%d",
				d.TotalCount, d.Page, d.TotalPages, len(d.Builds))
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		h.listTmpl.ExecuteTemplate(w, "content", d)
		return
	}
	h.listTmpl.Execute(w, d)
}

func (h *BuildHandler) Detail(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	num := chi.URLParam(r, "number")
	user := middleware.GetUser(r)
	partial := r.URL.Query().Get("partial") == "1"

	if !partial {
		selfURL := "/builds/" + owner + "/" + name + "/" + num
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(skeletonHTML(selfURL)))
			return
		}
		h.detailTmpl.Execute(w, map[string]interface{}{
			"User": user, "Loading": true, "SelfURL": selfURL,
		})
		return
	}

	var number int64
	fmt.Sscanf(num, "%d", &number)

	build, err := h.drone.GetBuild(owner, name, int(number))
	if err != nil {
		h.log.Error("[builds] GetBuild %s/%s #%d: %v", owner, name, number, err)
		http.Error(w, fmt.Sprintf("get build: %v", err), http.StatusNotFound)
		return
	}

	created := time.Unix(build.Created, 0).Format("2006-01-02 15:04")

	type logLine struct {
		Number  int
		Message string
	}
	type stepLogs struct {
		Number int
		Logs   []logLine
	}
	type stageLogs struct {
		Name  string
		Steps []stepLogs
	}

	var stages []stageLogs
	for _, st := range build.Stages {
		var steps []stepLogs
		for step := 1; ; step++ {
			lines, err := h.drone.GetBuildLogs(owner, name, int(number), st.Number, step)
			if err != nil || len(lines) == 0 {
				break
			}
			var sl []logLine
			for _, l := range lines {
				sl = append(sl, logLine{
					Number:  l.Number,
					Message: l.Message,
				})
			}
			steps = append(steps, stepLogs{Number: step, Logs: sl})
		}
		if len(steps) > 0 {
			stages = append(stages, stageLogs{Name: st.Name, Steps: steps})
		}
	}

	data := map[string]interface{}{
		"User":    user,
		"Owner":   owner,
		"Repo":    name,
		"Number":  build.Number,
		"Status":  build.Status,
		"Event":   build.Event,
		"Branch":  build.Target,
		"Author":  build.Author,
		"Created": created,
		"Stages":  stages,
	}

	if r.Header.Get("HX-Request") == "true" {
		h.detailTmpl.ExecuteTemplate(w, "content", data)
		return
	}
	h.detailTmpl.Execute(w, data)
}
