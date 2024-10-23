package backstageprocessor

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tdabasinskas/go-backstage/v2/backstage"
)

type backstageAPITransport struct {
	apiToken string
}

func (t *backstageAPITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Token "+t.apiToken)
	return http.DefaultTransport.RoundTrip(req)
}

type EntityWrapper struct {
	*backstage.Entity
}

type GithubRepoSpec struct {
	Implementation struct {
		Spec struct {
			Repository string `json:"repository"`
		} `json:"spec"`
	} `json:"implementation"`
}

type RepoInfo struct {
	Repo     string `json:"repo"`
	Org      string `json:"org"`
	Division string `json:"division"`
}

func getRepositoryLabelsMap(backstageUrl string, apiToken string) (map[string]RepoInfo, error) {
	entities, err := run(backstageUrl, apiToken, "kind=resource,spec.type=github-repository")
	if err != nil {
		return nil, err
	}
	repoMap := make(map[string]RepoInfo)
	for _, e := range entities {
		// we need to do a JSON round trip because the `e.Spec` type is `map[string]any`s all the way down. As we know exactly which fields we want, we can do the round trip to a `githubRepoSpec` and then pull the only fields we actually care about here
		b, err := json.Marshal(e.Spec)
		if err != nil {
			return nil, err
		}

		var spec GithubRepoSpec
		err = json.Unmarshal(b, &spec)
		if err != nil {
			return nil, err
		}

		// the service name uses the org - repo format
		// while repository in backstage uses org/repo format
		repoName := strings.ReplaceAll(spec.Implementation.Spec.Repository, "/", "-")
		repoInfo := RepoInfo{
			Repo:     repoName,
			Division: e.Metadata.Labels["division"],
			Org:      e.Metadata.Labels["org"],
		}

		repoMap[repoInfo.Repo] = repoInfo
	}

	return repoMap, nil
}

// run returns a list of entities based on the given condition
func run(backstageUrl string, apiToken string, filters string) ([]EntityWrapper, error) {
	httpClient := &http.Client{}
	httpClient.Transport = &backstageAPITransport{apiToken: apiToken}
	c, err := backstage.NewClient(backstageUrl, "default", httpClient)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	entities, _, err := c.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: []string{
			filters,
		},
		Order: []backstage.ListEntityOrder{{Direction: backstage.OrderAscending, Field: "metadata.name"}},
	})
	if err != nil {
		return nil, err
	}

	var wrappedEntities []EntityWrapper
	for _, entity := range entities {
		currentEntity := entity
		wrappedEntity := EntityWrapper{Entity: &currentEntity}
		wrappedEntities = append(wrappedEntities, wrappedEntity)
	}

	return wrappedEntities, nil
}
