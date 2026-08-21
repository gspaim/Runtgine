package playbooks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"gopkg.in/yaml.v3"
)

type Playbook struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Capabilities []string `yaml:"capabilities"`
	Body         string   `yaml:"-"`
}

func Dir(workspace string) string {
	return filepath.Join(workspace, ".runtgine", "playbooks")
}

func Load(dir string) ([]Playbook, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Playbook
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		pb, err := parse(raw)
		if err != nil || pb.ID == "" {
			continue
		}
		out = append(out, pb)
	}
	return out, nil
}

func parse(raw []byte) (Playbook, error) {
	text := string(raw)
	var pb Playbook
	if strings.HasPrefix(text, "---\n") {
		rest := text[4:]
		end := strings.Index(rest, "\n---")
		if end >= 0 {
			if err := yaml.Unmarshal([]byte(rest[:end]), &pb); err != nil {
				return Playbook{}, err
			}
			pb.Body = strings.TrimSpace(rest[end+4:])
			return pb, nil
		}
	}
	pb.Body = strings.TrimSpace(text)
	if pb.ID == "" {
		pb.ID = "untitled"
	}
	return pb, nil
}

func Hits(books []Playbook, capability string, maxHits, maxChars int) []contextpack.PlaybookHit {
	if maxHits <= 0 {
		maxHits = contextpack.DefaultPlaybookMaxHits
	}
	if maxChars <= 0 {
		maxChars = contextpack.DefaultPlaybookMaxChars
	}
	var hits []contextpack.PlaybookHit
	used := 0
	for _, b := range books {
		if !matches(b.Capabilities, capability) {
			continue
		}
		snip := b.Body
		if len(snip) > 400 {
			snip = snip[:400]
		}
		if used+len(snip) > maxChars {
			remain := maxChars - used
			if remain <= 0 {
				break
			}
			snip = snip[:remain]
		}
		hits = append(hits, contextpack.PlaybookHit{
			ID: b.ID, Title: b.Title, Snippet: snip,
		})
		used += len(snip)
		if len(hits) >= maxHits {
			break
		}
	}
	return hits
}

func matches(caps []string, capability string) bool {
	if len(caps) == 0 {
		return true
	}
	for _, c := range caps {
		if c == capability || strings.HasPrefix(capability, c) {
			return true
		}
	}
	return false
}
