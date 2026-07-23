package gitea

import (
	"encoding/base64"
	"fmt"
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

func (c *Client) ListContents(owner, name, ref, path string) ([]*gitea_sdk.ContentsResponse, error) {
	start := time.Now()
	entries, resp, err := c.sdk.ListContents(owner, name, ref, path)
	if err != nil {
		status := 0
		if resp != nil && resp.Response != nil {
			status = resp.Response.StatusCode
		}
		c.log.Info("[gitea] ListContents %s/%s ref=%s path=%s → http=%d err=%v (%v)", owner, name, ref, path, status, err, time.Since(start))
		return nil, err
	}
	c.log.Info("[gitea] ListContents %s/%s ref=%s path=%s → %d entries (%v)", owner, name, ref, path, len(entries), time.Since(start))
	return entries, nil
}

type FileResult struct {
	Content     string
	RawContent  string
	SHA         string
	LastSHA     string
	LastMessage string
	LastAuthor  string
	LastTime    string
	DownloadURL string
	Size        int64
}

func (c *Client) GetFile(owner, name, ref, path string) (*FileResult, error) {
	start := time.Now()
	resp, _, err := c.sdk.GetContents(owner, name, ref, path)
	if err != nil {
		c.log.Info("[gitea] GetContents %s/%s path=%s → %v (%v)", owner, name, path, err, time.Since(start))
		return nil, err
	}
	if resp.Content == nil {
		c.log.Info("[gitea] GetContents %s/%s path=%s → empty (%v)", owner, name, path, time.Since(start))
		return nil, fmt.Errorf("not a file")
	}
	data, err := base64.StdEncoding.DecodeString(*resp.Content)
	if err != nil {
		return nil, err
	}
	r := &FileResult{
		Content:    string(data),
		RawContent: string(data),
		SHA:        resp.SHA,
		Size:       resp.Size,
	}
	if resp.LastCommitSha != nil {
		r.LastSHA = *resp.LastCommitSha
	}
	if resp.LastCommitMessage != nil {
		r.LastMessage = *resp.LastCommitMessage
	}
	if resp.DownloadURL != nil {
		r.DownloadURL = *resp.DownloadURL
	}
	c.log.Info("[gitea] GetContents %s/%s path=%s → %d bytes sha=%s (%v)", owner, name, path, len(data), r.SHA, time.Since(start))
	return r, nil
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
