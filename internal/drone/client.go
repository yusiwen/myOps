package drone

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"

	applog "github.com/yusiwen/myops/internal/log"
)

type Client struct {
	client drone.Client
	log    *applog.Logger
}

func New(baseURL, token string, logger *applog.Logger) *Client {
	u, _ := url.Parse(baseURL)
	config := &oauth2.Config{}
	auther := config.Client(context.Background(), &oauth2.Token{AccessToken: token})
	httpClient := &http.Client{
		Transport: auther.Transport,
	}
	client := drone.NewClient(u.String(), httpClient)
	return &Client{client: client, log: logger}
}

func (c *Client) ListRepos() ([]*drone.Repo, error) {
	start := time.Now()
	repos, err := c.client.RepoList()
	if err != nil {
		c.log.Error("[drone] RepoList failed: %v", err)
		return nil, err
	}
	c.log.Info("[drone] RepoList → %d repos (%v)", len(repos), time.Since(start))
	return repos, nil
}

func (c *Client) ListBuilds(owner, name string) ([]*drone.Build, error) {
	start := time.Now()
	builds, err := c.client.BuildList(owner, name, drone.ListOptions{})
	if err != nil {
		c.log.Error("[drone] BuildList %s/%s failed: %v", owner, name, err)
		return nil, err
	}
	c.log.Info("[drone] BuildList %s/%s → %d builds (%v)", owner, name, len(builds), time.Since(start))
	return builds, nil
}

func (c *Client) GetBuild(owner, name string, number int) (*drone.Build, error) {
	start := time.Now()
	build, err := c.client.Build(owner, name, number)
	if err != nil {
		c.log.Error("[drone] Build %s/%s #%d failed: %v", owner, name, number, err)
		return nil, err
	}
	c.log.Info("[drone] Build %s/%s #%d → status=%s (%v)", owner, name, number, build.Status, time.Since(start))
	return build, nil
}

func (c *Client) GetBuildLogs(owner, name string, build, stage, step int) ([]*drone.Line, error) {
	start := time.Now()
	logs, err := c.client.Logs(owner, name, build, stage, step)
	if err != nil {
		c.log.Info("[drone] Logs %s/%s #%d(%d/%d) → empty (%v)", owner, name, build, stage, step, time.Since(start))
		return nil, nil
	}
	c.log.Info("[drone] Logs %s/%s #%d(%d/%d) → %d lines (%v)", owner, name, build, stage, step, len(logs), time.Since(start))
	return logs, nil
}

func (c *Client) BuildLast(owner, name string) (*drone.Build, error) {
	start := time.Now()
	build, err := c.client.BuildLast(owner, name, "")
	if err != nil {
		c.log.Info("[drone] BuildLast %s/%s → empty (%v)", owner, name, time.Since(start))
		return nil, nil
	}
	c.log.Info("[drone] BuildLast %s/%s → #%d status=%s (%v)", owner, name, build.Number, build.Status, time.Since(start))
	return build, nil
}

func (c *Client) GetRepo(owner, name string) (*drone.Repo, error) {
	start := time.Now()
	repo, err := c.client.Repo(owner, name)
	if err != nil {
		c.log.Error("[drone] Repo %s/%s failed: %v", owner, name, err)
		return nil, err
	}
	c.log.Info("[drone] Repo %s/%s → ok (%v)", owner, name, time.Since(start))
	return repo, nil
}
