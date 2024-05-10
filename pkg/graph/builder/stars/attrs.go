package stars

import (
	"github.com/google/go-github/v61/github"
)

func OwnerAttrs(owner *github.User) map[string]interface{} {
	attrs := map[string]interface{}{
		"name":              owner.Login,
		"type":              owner.Type,
		"url":               owner.URL,
		"html_url":          owner.HTMLURL,
		"repos_url":         owner.ReposURL,
		"followers":         owner.Followers,
		"followers_url":     owner.FollowersURL,
		"following":         owner.Following,
		"following_url":     owner.FollowingURL,
		"starred_url":       owner.StarredURL,
		"created_at":        owner.CreatedAt,
		"updated_at":        owner.UpdatedAt,
		"collaborators":     owner.Collaborators,
		"organizations_url": owner.OrganizationsURL,
	}
	return attrs
}

func RepoAttrs(repo *github.Repository, starredAt *github.Timestamp) map[string]interface{} {
	attrs := map[string]interface{}{
		"name":              repo.Name,
		"full_name":         repo.FullName,
		"visibility":        repo.Visibility,
		"archived":          repo.Archived,
		"starred_at":        starredAt,
		"created_at":        repo.CreatedAt,
		"pushed_at":         repo.PushedAt,
		"updated_at":        repo.UpdatedAt,
		"html_url":          repo.HTMLURL,
		"forks_url":         repo.ForksURL,
		"languages_url":     repo.LanguagesURL,
		"collaborators_url": repo.CollaboratorsURL,
		"stargazers_count":  repo.StargazersCount,
		"stargazers_url":    repo.StargazersURL,
	}
	if repo.License != nil {
		attrs["license"] = repo.License.Name
	}
	return attrs
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
		"url":  "https://github.com/trending/" + lang,
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
