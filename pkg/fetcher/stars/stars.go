package stars

import (
	"context"
	"fmt"

	"github.com/google/go-github/v61/github"
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

func (f *Fetcher) GetTotalPages(ctx context.Context, paging int) (int, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: f.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{
			PerPage: paging,
		},
	}

	_, resp, err := client.Activity.ListStarred(ctx, f.user, opts)
	if err != nil {
		return 0, err
	}

	return resp.LastPage, nil
}

func (f *Fetcher) Fetch(ctx context.Context, startPage, endPage int, reposChan chan<- interface{}) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: f.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	for page := startPage; page <= endPage; page++ {
		opts := &github.ActivityListStarredOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: f.paging,
			},
		}

		repos, _, err := client.Activity.ListStarred(ctx, "", opts)
		if err != nil {
			return fmt.Errorf("error fetching page %d: %v", page, err)
		}

		select {
		case reposChan <- repos:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
