package main

import (
	"context"
	"testing"

	"github.com/wantedly/webmock-proxy/webmock"
)

func TestApiRequest(t *testing.T) {
	px, err := webmock.NewProxy()
	if err != nil {
		t.Fatal(err)
	}
	defer px.Close()
	github := &GithubAPI{Client: px.Client}
	repo, err := github.RepoInfo(context.Background(), "wantedly", "webmock-proxy")
	if err != nil {
		t.Fatal(err)
	}
	if repo.Name != "webmock-proxy" {
		t.Errorf(`expected repository name is "webmock-proxy", but got %q`, repo.Name)
	}
	if repo.FullName != "wantedly/webmock-proxy" {
		t.Errorf(`expected repository full name is "wantedly/webmock-proxy", but got %q`, repo.FullName)
	}
	if repo.HTMLURL != "https://github.com/wantedly/webmock-proxy" {
		t.Errorf(`expected repository html url is "https://github.com/wantedly/webmock-proxy", but got %q`, repo.HTMLURL)
	}
}
