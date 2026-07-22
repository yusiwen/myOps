package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yusiwen/myops/internal/code"
	"github.com/yusiwen/myops/internal/gitea"
	applog "github.com/yusiwen/myops/internal/log"
	"github.com/yusiwen/myops/internal/middleware"
)

const reposPerPage = 50

type RepoHandler struct {
	gitea      *gitea.Client
	render     code.CodeRenderer
	log        *applog.Logger
	listTmpl   *template.Template
	detailTmpl *template.Template
}

func NewRepoHandler(gitea *gitea.Client, render code.CodeRenderer, logger *applog.Logger) *RepoHandler {
	return &RepoHandler{
		gitea:      gitea,
		render:     render,
		log:        logger,
		listTmpl:   template.Must(template.ParseFiles("web/templates/base.html", "web/templates/repo_list.html")),
		detailTmpl: template.Must(template.ParseFiles("web/templates/base.html", "web/templates/repo_detail.html")),
	}
}

type repoItem struct {
	Name       string
	Owner      string
	Visibility string
	Language   string
	CloneURL   string
}

type repoListData struct {
	User       *middleware.SessionUser
	Repos      []repoItem
	DroneError string
	Query      string
	Page       int
	PrevPage   int
	NextPage   int
	TotalPages int
	TotalCount int
}

func (h *RepoHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	partial := r.URL.Query().Get("partial") == "1"

	if !partial {
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(skeletonHTML("/repos")))
			return
		}
		h.listTmpl.Execute(w, map[string]interface{}{
			"User": user, "Loading": true, "SelfURL": "/repos",
		})
		return
	}

	q := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	d := repoListData{User: user, Query: q, Page: page}
	if page > 1 {
		d.PrevPage = page - 1
	}
	d.NextPage = page + 1

	if h.gitea != nil {
		if q != "" {
			repos, total, err := h.gitea.SearchRepos(q, page, reposPerPage)
			if err != nil {
				h.log.Error("[repos] SearchRepos: %v", err)
				d.DroneError = err.Error()
			} else {
				d.TotalCount = total
				d.TotalPages = (total + reposPerPage - 1) / reposPerPage
				for _, r := range repos {
					vis := "public"
					if r.Private {
						vis = "private"
					}
					owner := ""
					if r.Owner != nil {
						owner = r.Owner.UserName
					}
					d.Repos = append(d.Repos, repoItem{
						Name:       r.Name,
						Owner:      owner,
						Visibility: vis,
						Language:   r.Language,
						CloneURL:   r.CloneURL,
					})
				}
			}
		} else {
			list, err := h.gitea.ListRepos()
			if err != nil {
				h.log.Error("[repos] ListRepos: %v", err)
				d.DroneError = err.Error()
			} else {
				d.TotalCount = len(list)
				d.TotalPages = (d.TotalCount + reposPerPage - 1) / reposPerPage
				start := (page - 1) * reposPerPage
				end := start + reposPerPage
				if end > d.TotalCount {
					end = d.TotalCount
				}
				if start < d.TotalCount {
					for _, r := range list[start:end] {
						vis := "public"
						if r.Private {
							vis = "private"
						}
						owner := ""
						if r.Owner != nil {
							owner = r.Owner.UserName
						}
						d.Repos = append(d.Repos, repoItem{
							Name:       r.Name,
							Owner:      owner,
							Visibility: vis,
							Language:   r.Language,
							CloneURL:   r.CloneURL,
						})
					}
				}
			}
		}
	}

	h.listTmpl.ExecuteTemplate(w, "content", d)
}

func (h *RepoHandler) Detail(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	user := middleware.GetUser(r)
	partial := r.URL.Query().Get("partial") == "1"

	if !partial {
		selfURL := "/repos/" + owner + "/" + name
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(skeletonHTML(selfURL)))
			return
		}
		h.detailTmpl.Execute(w, map[string]interface{}{
			"User": user, "Loading": true, "SelfURL": selfURL,
		})
		return
	}

	repo, err := h.gitea.GetRepo(owner, name)
	if err != nil {
		h.log.Error("[repos] GetRepo %s/%s: %v", owner, name, err)
		http.Error(w, fmt.Sprintf("get repo: %v", err), http.StatusInternalServerError)
		return
	}

	type fileItem struct {
		Name   string
		Size   string
		IsDir  bool
		Svg    string
	}

	vis := "public"
	if repo.Private {
		vis = "private"
	}

	var files []fileItem
	var fileErr string
	if repo.Empty {
		fileErr = "Empty repository"
	} else {
		entries, err := h.gitea.ListContents(owner, name, "", "/")
		if err != nil {
			fileErr = err.Error()
			h.log.Error("[repos] ListContents %s/%s: %v", owner, name, err)
		} else {
			for _, e := range entries {
				files = append(files, fileItem{
					Name:  e.Name,
					Size:  fmtSize(e.Size),
					IsDir: e.Type == "dir",
				})
			}
		}
	}

	data := map[string]interface{}{
		"User":          user,
		"Owner":         owner,
		"Name":          name,
		"Description":   repo.Description,
		"DefaultBranch": repo.DefaultBranch,
		"Visibility":    vis,
		"Language":      repo.Language,
		"CloneURL":      repo.CloneURL,
		"Files":         files,
		"FileError":     fileErr,
	}
	h.detailTmpl.ExecuteTemplate(w, "content", data)
}

func (h *RepoHandler) File(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	name := chi.URLParam(r, "name")
	filepath := r.URL.Query().Get("path")

	repo, err := h.gitea.GetRepo(owner, name)
	if err != nil {
		http.Error(w, "repo not found", http.StatusNotFound)
		return
	}

	_ = repo

	content := fmt.Sprintf("file: %s/%s/%s\n\n(content placeholder)", owner, name, filepath)
	highlighted := strings.Builder{}
	formatted, err := h.render.Render(filepath, []byte(content))
	if err == nil {
		highlighted.WriteString(string(formatted.HTML))
	} else {
		highlighted.WriteString(string(content))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(highlighted.String()))
}

func fmtSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.0f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}
