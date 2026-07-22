package gitea

import (
	"net/url"
	"strconv"
	"time"

	gitea_sdk "code.gitea.io/sdk/gitea"

	applog "github.com/yusiwen/myops/internal/log"
)

type Client struct {
	baseURL string
	sdk     *gitea_sdk.Client
	log     *applog.Logger
}

func New(baseURL, token string, logger *applog.Logger) *Client {
	u, _ := url.Parse(baseURL)
	sdk, err := gitea_sdk.NewClient(u.String(),
		gitea_sdk.SetToken(token),
		gitea_sdk.SetGiteaVersion(""),
	)
	if err != nil {
		logger.Error("[gitea] init failed: %v", err)
		return nil
	}
	return &Client{baseURL: u.String(), sdk: sdk, log: logger}
}

func (c *Client) GetUser(token string) (*gitea_sdk.User, error) {
	start := time.Now()
	client, err := gitea_sdk.NewClient(c.baseURL, gitea_sdk.SetToken(token))
	if err != nil {
		c.log.Error("[gitea] GetUser failed: %v", err)
		return nil, err
	}
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		c.log.Error("[gitea] GetMyUserInfo failed: %v", err)
		return nil, err
	}
	c.log.Info("[gitea] GetMyUserInfo → %s (%v)", user.UserName, time.Since(start))
	return user, nil
}

func (c *Client) RepoCount() (int, error) {
	start := time.Now()
	repos, resp, err := c.sdk.ListMyRepos(gitea_sdk.ListReposOptions{
		ListOptions: gitea_sdk.ListOptions{PageSize: 1},
	})
	if err != nil {
		c.log.Error("[gitea] RepoCount failed: %v", err)
		return 0, err
	}
	total := len(repos)
	if resp != nil && resp.Header != nil {
		if v := resp.Header.Get("X-Total-Count"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				total = n
			}
		}
	}
	c.log.Info("[gitea] RepoCount → %d (%v)", total, time.Since(start))
	return total, nil
}

func (c *Client) ListRepos() ([]*gitea_sdk.Repository, error) {
	start := time.Now()
	repos, _, err := c.sdk.ListMyRepos(gitea_sdk.ListReposOptions{
		ListOptions: gitea_sdk.ListOptions{PageSize: 200},
	})
	if err != nil {
		c.log.Error("[gitea] ListRepos failed: %v", err)
		return nil, err
	}
	c.log.Info("[gitea] ListRepos → %d repos (%v)", len(repos), time.Since(start))
	return repos, nil
}

func (c *Client) SearchRepos(keyword string, page, pageSize int) ([]*gitea_sdk.Repository, int, error) {
	start := time.Now()
	opts := gitea_sdk.SearchRepoOptions{
		Keyword:    keyword,
		ListOptions: gitea_sdk.ListOptions{Page: page, PageSize: pageSize},
	}
	repos, resp, err := c.sdk.SearchRepos(opts)
	if err != nil {
		c.log.Error("[gitea] SearchRepos q=%q failed: %v", keyword, err)
		return nil, 0, err
	}
	total := len(repos)
	if resp != nil && resp.Header != nil {
		if v := resp.Header.Get("X-Total-Count"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				total = n
			}
		}
	}
	c.log.Info("[gitea] SearchRepos q=%q page=%d → %d/%d (%v)", keyword, page, len(repos), total, time.Since(start))
	return repos, total, nil
}

func (c *Client) GetRepo(owner, name string) (*gitea_sdk.Repository, error) {
	start := time.Now()
	repo, _, err := c.sdk.GetRepo(owner, name)
	if err != nil {
		c.log.Error("[gitea] GetRepo %s/%s failed: %v", owner, name, err)
		return nil, err
	}
	c.log.Info("[gitea] GetRepo %s/%s → ok (%v)", owner, name, time.Since(start))
	return repo, nil
}
