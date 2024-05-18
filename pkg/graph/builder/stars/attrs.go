package stars

import (
	"github.com/google/go-github/v61/github"
	"github.com/milosgajdos/orbnet/pkg/graph/attrs"
)

func OwnerAttrs(owner *github.User) (map[string]interface{}, error) {
	return attrs.Encode(owner)
}

func RepoAttrs(repo *github.Repository, starredAt *github.Timestamp) (map[string]interface{}, error) {
	attrs, err := attrs.Encode(repo)
	if err != nil {
		return nil, err
	}
	attrs["starred_at"] = starredAt
	return attrs, nil
}

func TopicAttrs(topic string) map[string]interface{} {
	attrs := map[string]interface{}{
		"name": topic,
		"url":  "https://github.com/topics/" + topic,
	}
	return attrs
}

func LangAttrs(lang string) map[string]interface{} {
	attrs := map[string]interface{}{
		"name": lang,
	}
	return attrs
}

func LinkAttrs(rel string, weight float64) map[string]interface{} {
	attrs := map[string]interface{}{
		"relation": rel,
		"weight":   weight,
	}
	return attrs
}
