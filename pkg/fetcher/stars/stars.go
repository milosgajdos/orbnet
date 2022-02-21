package stars

import (
	"context"

	"github.com/google/go-github/v34/github"
	"golang.org/x/oauth2"
)

type Fetcher struct {
	user   string
	token  string
	paging int
}

func NewFetcher(token, user string, paging int) (*Fetcher, error) {
	return &Fetcher{
		user:   user,
		token:  token,
		paging: paging,
	}, nil
}

func (f *Fetcher) Fetch(ctx context.Context, reposChan chan<- interface{}) error {
	defer close(reposChan)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: f.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: f.paging},
	}

	for {
		repos, resp, err := client.Activity.ListStarred(ctx, f.user, opts)
		if err != nil {
			return err
		}

		select {
		case reposChan <- repos:
		case <-ctx.Done():
			return nil
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return nil
}
