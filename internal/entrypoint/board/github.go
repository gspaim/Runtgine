package board

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Card struct {
	ID     string
	Title  string
	Body   string
	Repo   string // owner/name
	Number int    // issue number for write-back
	URL    string
}

type GitHub struct {
	Token  string
	Client *http.Client
}

func NewGitHub(token string) *GitHub {
	return &GitHub{Token: token, Client: &http.Client{Timeout: 30 * time.Second}}
}

func (g *GitHub) ListIssueCards(ctx context.Context, repo, label string) ([]Card, error) {
	if repo == "" {
		return nil, fmt.Errorf("repo is required (owner/name)")
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=30", repo)
	if label != "" {
		u += "&labels=" + label
	}
	b, err := g.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var issues []struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		PullReq *any   `json:"pull_request"`
	}
	if err := json.Unmarshal(b, &issues); err != nil {
		return nil, err
	}
	var cards []Card
	for _, is := range issues {
		if is.PullReq != nil {
			continue
		}
		cards = append(cards, Card{
			ID:     fmt.Sprintf("%s#%d", repo, is.Number),
			Title:  is.Title,
			Body:   is.Body,
			Repo:   repo,
			Number: is.Number,
			URL:    is.HTMLURL,
		})
	}
	return cards, nil
}

func (g *GitHub) ListProjectItems(ctx context.Context, owner string, projectNumber int, ownerIsOrg bool) ([]Card, error) {
	field := "user"
	if ownerIsOrg {
		field = "organization"
	}
	query := fmt.Sprintf(`query($login:String!, $n:Int!) {
  %s(login:$login) {
    projectV2(number:$n) {
      items(first:30) {
        nodes {
          id
          content {
            ... on Issue { number title body url repository { nameWithOwner } }
          }
        }
      }
    }
  }
}`, field)
	payload, _ := json.Marshal(map[string]any{
		"query": query,
		"variables": map[string]any{"login": owner, "n": projectNumber},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	g.headers(req)
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github graphql %d: %s", resp.StatusCode, truncate(string(b), 300))
	}
	var parsed struct {
		Data map[string]struct {
			ProjectV2 struct {
				Items struct {
					Nodes []struct {
						ID      string `json:"id"`
						Content *struct {
							Number     int    `json:"number"`
							Title      string `json:"title"`
							Body       string `json:"body"`
							URL        string `json:"url"`
							Repository struct {
								NameWithOwner string `json:"nameWithOwner"`
							} `json:"repository"`
						} `json:"content"`
					} `json:"nodes"`
				} `json:"items"`
			} `json:"projectV2"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("github graphql: %s", parsed.Errors[0].Message)
	}
	block := parsed.Data[field]
	var cards []Card
	for _, n := range block.ProjectV2.Items.Nodes {
		if n.Content == nil || n.Content.Title == "" {
			continue
		}
		cards = append(cards, Card{
			ID:     n.ID,
			Title:  n.Content.Title,
			Body:   n.Content.Body,
			Repo:   n.Content.Repository.NameWithOwner,
			Number: n.Content.Number,
			URL:    n.Content.URL,
		})
	}
	return cards, nil
}

func (g *GitHub) Comment(ctx context.Context, repo string, number int, body string) error {
	if repo == "" || number == 0 {
		return nil
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments", repo, number)
	payload, _ := json.Marshal(map[string]string{"body": body})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	g.headers(req)
	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github comment %d: %s", resp.StatusCode, truncate(string(b), 300))
	}
	return nil
}

func (g *GitHub) AddLabel(ctx context.Context, repo string, number int, label string) error {
	if repo == "" || number == 0 || label == "" {
		return nil
	}
	u := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/labels", repo, number)
	payload, _ := json.Marshal([]string{label})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	g.headers(req)
	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github label %d: %s", resp.StatusCode, truncate(string(b), 200))
	}
	return nil
}

func (g *GitHub) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	g.headers(req)
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github %d: %s", resp.StatusCode, truncate(string(b), 300))
	}
	return b, nil
}

func (g *GitHub) headers(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}
