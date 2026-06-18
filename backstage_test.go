package backstageprocessor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("successfully lists and wraps entities", func(t *testing.T) {
		var authHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader = r.Header.Get("Authorization")

			if r.URL.Path != "/api/catalog/entities" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"apiVersion": "backstage.io/v1alpha1",
					"kind": "Resource",
					"metadata": {
						"name": "repo-one",
						"labels": {"org": "engineering", "division": "platform"}
					},
					"spec": {
						"implementation": {
							"spec": {
								"repository": "engineering/repo-one"
							}
						}
					}
				},
				{
					"apiVersion": "backstage.io/v1alpha1",
					"kind": "Resource",
					"metadata": {
						"name": "repo-two",
						"labels": {"org": "engineering", "division": "data"}
					},
					"spec": {
						"implementation": {
							"spec": {
								"repository": "engineering/repo-two"
							}
						}
					}
				}
			]`))
		}))
		defer server.Close()

		entities, err := run(server.URL, "secret-token", "kind=resource,spec.type=github-repository")
		if err != nil {
			t.Fatalf("run() returned error: %v", err)
		}

		if authHeader != "Token secret-token" {
			t.Fatalf("expected Authorization header to be set, got %q", authHeader)
		}

		if len(entities) != 2 {
			t.Fatalf("expected 2 entities, got %d", len(entities))
		}

		if entities[0].Metadata.Name != "repo-one" {
			t.Fatalf("unexpected first entity name: %s", entities[0].Metadata.Name)
		}
		if entities[1].Metadata.Name != "repo-two" {
			t.Fatalf("unexpected second entity name: %s", entities[1].Metadata.Name)
		}
	})

	t.Run("returns error on invalid client URL", func(t *testing.T) {
		_, err := run("not a valid url", "token", "kind=resource")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})

	t.Run("returns error on non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
		}))
		defer server.Close()

		_, err := run(server.URL, "token", "kind=resource")
		if err == nil {
			t.Fatal("expected error when backend returns non-2xx")
		}
	})
}

func TestGetRepositoryLabelsMap(t *testing.T) {
	t.Run("maps repository labels from entities", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/catalog/entities" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"apiVersion": "backstage.io/v1alpha1",
					"kind": "Resource",
					"metadata": {
						"name": "repo-one",
						"labels": {"org": "engineering", "division": "platform"}
					},
					"spec": {
						"implementation": {
							"spec": {
								"repository": "engineering/repo-one"
							}
						}
					}
				}
			]`))
		}))
		defer server.Close()

		repos, err := getRepositoryLabelsMap(server.URL, "token")
		if err != nil {
			t.Fatalf("getRepositoryLabelsMap() returned error: %v", err)
		}

		if len(repos) != 1 {
			t.Fatalf("expected exactly one repository, got %d", len(repos))
		}

		repo, ok := repos["engineering-repo-one"]
		if !ok {
			t.Fatal("expected key engineering-repo-one to exist")
		}

		if repo.Org != "engineering" {
			t.Fatalf("expected org engineering, got %s", repo.Org)
		}
		if repo.Division != "platform" {
			t.Fatalf("expected division platform, got %s", repo.Division)
		}
	})

	t.Run("returns unmarshal error for unexpected repository type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"apiVersion": "backstage.io/v1alpha1",
					"kind": "Resource",
					"metadata": {
						"name": "broken-repo",
						"labels": {"org": "engineering", "division": "platform"}
					},
					"spec": {
						"implementation": {
							"spec": {
								"repository": 123
							}
						}
					}
				}
			]`))
		}))
		defer server.Close()

		_, err := getRepositoryLabelsMap(server.URL, "token")
		if err == nil {
			t.Fatal("expected unmarshal error for non-string repository")
		}
	})

	t.Run("returns error when listing entities fails", func(t *testing.T) {
		_, err := getRepositoryLabelsMap("not a valid url", "token")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}
