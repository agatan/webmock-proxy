package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/pkg/errors"
)

var githubAPIURL *url.URL

func init() {
	var err error
	githubAPIURL, err = url.ParseRequestURI("https://api.github.com/")
	if err != nil {
		panic(fmt.Sprintf("failed to parse github api request: %v", err))
	}
}

type GithubAPI struct {
	Client *http.Client
	URL    *url.URL
}

func (g *GithubAPI) client() *http.Client {
	if g.Client == nil {
		return http.DefaultClient
	}
	return g.Client
}

func (g *GithubAPI) newRequest(ctx context.Context, method, rpath string, body io.Reader) (*http.Request, error) {
	var url url.URL
	if g.URL == nil {
		url = *githubAPIURL
	} else {
		url = *g.URL
	}
	url.Path = path.Join(url.Path, rpath)

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	// configure any additional information like custom headers or basic auth.

	return req, nil
}

func (g *GithubAPI) decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decorder := json.NewDecoder(resp.Body)
	return decorder.Decode(out)
}

type RepoInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	URL      string `json:"url"`
	// some fields are omitted...
}

func (g *GithubAPI) RepoInfo(ctx context.Context, owner string, name string) (*RepoInfo, error) {
	req, err := g.newRequest(ctx, "GET", fmt.Sprintf("/repos/%s/%s", owner, name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := g.client().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("failed to get repository info with status: %s", resp.Status)
	}
	var repo RepoInfo
	if err := g.decodeBody(resp, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

func main() {
	var g GithubAPI
	repo, err := g.RepoInfo(context.Background(), "wantedly", "webmock-proxy")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	fmt.Printf("%#v\n", repo)
}
