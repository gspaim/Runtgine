package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/gspaim/Runtgine/internal/core/registry"
	"github.com/gspaim/Runtgine/internal/core/result"
)

const (
	CapRead  = "fs.read"
	CapWrite = "fs.write"
	CapList  = "fs.list"
	CapStat  = "fs.stat"

	defaultMaxBytes   = 1 << 20
	maxBytesLimit     = 4 << 20
	defaultMaxEntries = 200
	maxEntriesLimit   = 1000
)

type Player struct{}

func New() *Player { return &Player{} }

func (p *Player) Manifest() registry.Manifest {
	return registry.Manifest{
		SchemaVersion: "0.1.0",
		Name:          "filesystem",
		Version:       "0.1.0",
		Kind:          registry.KindDeterministic,
		Capabilities: []registry.Capability{
			{
				Name: CapRead,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path"],
  "properties":{
    "path":{"type":"string","minLength":1},
    "max_bytes":{"type":"integer","minimum":1,"maximum":4194304}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path","content","bytes","truncated"],
  "properties":{
    "path":{"type":"string"},
    "content":{"type":"string"},
    "bytes":{"type":"integer"},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapWrite,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path","content"],
  "properties":{
    "path":{"type":"string","minLength":1},
    "content":{"type":"string"},
    "create_parents":{"type":"boolean","default":false}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path","bytes","created"],
  "properties":{
    "path":{"type":"string"},
    "bytes":{"type":"integer"},
    "created":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapList,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "properties":{
    "path":{"type":"string","default":"."},
    "recursive":{"type":"boolean","default":false},
    "max_entries":{"type":"integer","minimum":1,"maximum":1000}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path","entries","truncated"],
  "properties":{
    "path":{"type":"string"},
    "entries":{"type":"array","items":{
      "type":"object",
      "required":["path","type"],
      "properties":{
        "path":{"type":"string"},
        "type":{"type":"string","enum":["file","directory","symlink"]}
      }
    }},
    "truncated":{"type":"boolean"}
  }
}`),
			},
			{
				Name: CapStat,
				InputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path"],
  "properties":{
    "path":{"type":"string","minLength":1}
  },
  "additionalProperties":false
}`),
				OutputSchema: json.RawMessage(`{
  "type":"object",
  "required":["path","type","size","mode","modified_at"],
  "properties":{
    "path":{"type":"string"},
    "type":{"type":"string","enum":["file","directory","symlink"]},
    "size":{"type":"integer"},
    "mode":{"type":"string"},
    "modified_at":{"type":"string"}
  }
}`),
			},
		},
	}
}

type readIn struct {
	Path     string `json:"path"`
	MaxBytes int    `json:"max_bytes"`
}

type readOut struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
}

type writeIn struct {
	Path          string `json:"path"`
	Content       string `json:"content"`
	CreateParents bool   `json:"create_parents"`
}

type writeOut struct {
	Path    string `json:"path"`
	Bytes   int    `json:"bytes"`
	Created bool   `json:"created"`
}

type listIn struct {
	Path       string `json:"path"`
	Recursive  bool   `json:"recursive"`
	MaxEntries int    `json:"max_entries"`
}

type listEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type listOut struct {
	Path      string      `json:"path"`
	Entries   []listEntry `json:"entries"`
	Truncated bool        `json:"truncated"`
}

type statIn struct {
	Path string `json:"path"`
}

type statOut struct {
	Path       string `json:"path"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	Mode       string `json:"mode"`
	ModifiedAt string `json:"modified_at"`
}

// ValidateStaticInput applies confinement and limit checks before admission (G-77/G-78).
func ValidateStaticInput(workspace, capability string, raw json.RawMessage) error {
	switch capability {
	case CapRead:
		var in readIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid fs.read input: "+err.Error(), nil)
		}
		if _, err := clampBytes(in.MaxBytes); err != nil {
			return err
		}
		_, _, err := resolvePath(workspace, in.Path, resolveOpts{allowRoot: false})
		return err
	case CapWrite:
		var in writeIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid fs.write input: "+err.Error(), nil)
		}
		if err := validateWriteContent(in.Content); err != nil {
			return err
		}
		_, _, err := resolvePath(workspace, in.Path, resolveOpts{
			allowMissing:  true,
			forWrite:      true,
			createParents: in.CreateParents,
		})
		return err
	case CapList:
		var in listIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid fs.list input: "+err.Error(), nil)
		}
		if _, err := clampEntries(in.MaxEntries); err != nil {
			return err
		}
		path := in.Path
		if path == "" {
			path = "."
		}
		_, _, err := resolvePath(workspace, path, resolveOpts{allowRoot: true})
		return err
	case CapStat:
		var in statIn
		if err := json.Unmarshal(raw, &in); err != nil {
			return result.Validation(result.CodeInvalidInput, "invalid fs.stat input: "+err.Error(), nil)
		}
		_, _, err := resolvePath(workspace, in.Path, resolveOpts{allowRoot: true})
		return err
	default:
		return result.Validation(result.CodeUnknownCapability, "filesystem player cannot validate "+capability, nil)
	}
}

func (p *Player) Execute(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, result.Runtime(result.CodeCancelled, "filesystem cancelled", false, nil)
	}
	if err := ValidateStaticInput(req.Workspace, req.Capability, req.Input); err != nil {
		return nil, err
	}
	switch req.Capability {
	case CapRead:
		return execRead(req)
	case CapWrite:
		return execWrite(req)
	case CapList:
		return execList(ctx, req)
	case CapStat:
		return execStat(req)
	default:
		return nil, result.Validation(result.CodeUnknownCapability, "filesystem player cannot serve "+req.Capability, nil)
	}
}

func execRead(req registry.ExecRequest) (json.RawMessage, error) {
	var in readIn
	_ = json.Unmarshal(req.Input, &in)
	limit, err := clampBytes(in.MaxBytes)
	if err != nil {
		return nil, err
	}
	abs, rel, err := resolvePath(req.Workspace, in.Path, resolveOpts{})
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "fs.read path not found", map[string]any{"path": rel})
	}
	if err := rejectEscapingSymlink(req.Workspace, abs, info); err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := confinedSymlinkTarget(req.Workspace, abs)
		if err != nil {
			return nil, err
		}
		info, err = os.Stat(target)
		if err != nil {
			return nil, result.Validation(result.CodeInvalidInput, "fs.read symlink target not found", map[string]any{"path": rel})
		}
		abs = target
	}
	if !info.Mode().IsRegular() {
		return nil, result.Validation(result.CodeInvalidInput, "fs.read requires a regular file", map[string]any{"path": rel})
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.read open: "+err.Error(), false, nil)
	}
	defer f.Close()
	buf := make([]byte, limit+1)
	n, err := io.ReadFull(f, buf)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil
	}
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.read: "+err.Error(), false, nil)
	}
	truncated := n > limit
	if truncated {
		n = limit
	}
	data := buf[:n]
	if truncated {
		for len(data) > 0 {
			r, size := utf8.DecodeLastRune(data)
			if r != utf8.RuneError || size != 1 {
				break
			}
			data = data[:len(data)-1]
		}
	}
	if !utf8.Valid(data) {
		return nil, result.Validation(result.CodeInvalidInput, "fs.read content is not valid UTF-8", map[string]any{"path": rel})
	}
	return json.Marshal(readOut{
		Path:      rel,
		Content:   string(data),
		Bytes:     len(data),
		Truncated: truncated,
	})
}

func execWrite(req registry.ExecRequest) (json.RawMessage, error) {
	var in writeIn
	_ = json.Unmarshal(req.Input, &in)
	if err := validateWriteContent(in.Content); err != nil {
		return nil, err
	}
	abs, rel, err := resolvePath(req.Workspace, in.Path, resolveOpts{
		allowMissing:  true,
		forWrite:      true,
		createParents: in.CreateParents,
	})
	if err != nil {
		return nil, err
	}
	created := true
	if info, err := os.Lstat(abs); err == nil {
		created = false
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, result.Validation(result.CodeInvalidInput, "fs.write destination must not be a symlink", map[string]any{"path": rel})
		}
		if info.IsDir() {
			return nil, result.Validation(result.CodeInvalidInput, "fs.write destination must not be a directory", map[string]any{"path": rel})
		}
	} else if !os.IsNotExist(err) {
		return nil, result.Runtime(result.CodePlayerError, "fs.write stat: "+err.Error(), false, nil)
	}

	parent := filepath.Dir(abs)
	if in.CreateParents {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return nil, result.Runtime(result.CodePlayerError, "fs.write create_parents: "+err.Error(), false, nil)
		}
	}
	tmp, err := os.CreateTemp(parent, ".runtgine-fs-write-*")
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.write tempfile: "+err.Error(), false, nil)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		_ = tmp.Close()
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.WriteString(in.Content); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.write: "+err.Error(), false, nil)
	}
	if err := tmp.Sync(); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.write sync: "+err.Error(), false, nil)
	}
	if err := tmp.Close(); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.write close: "+err.Error(), false, nil)
	}
	if err := os.Rename(tmpName, abs); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "fs.write rename: "+err.Error(), false, nil)
	}
	cleanup = false
	return json.Marshal(writeOut{
		Path:    rel,
		Bytes:   len(in.Content),
		Created: created,
	})
}

func execList(ctx context.Context, req registry.ExecRequest) (json.RawMessage, error) {
	var in listIn
	_ = json.Unmarshal(req.Input, &in)
	limit, err := clampEntries(in.MaxEntries)
	if err != nil {
		return nil, err
	}
	path := in.Path
	if path == "" {
		path = "."
	}
	abs, rel, err := resolvePath(req.Workspace, path, resolveOpts{allowRoot: true})
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "fs.list path not found", map[string]any{"path": rel})
	}
	if err := rejectEscapingSymlink(req.Workspace, abs, info); err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := confinedSymlinkTarget(req.Workspace, abs)
		if err != nil {
			return nil, err
		}
		info, err = os.Stat(target)
		if err != nil {
			return nil, result.Validation(result.CodeInvalidInput, "fs.list symlink target not found", map[string]any{"path": rel})
		}
		abs = target
	}
	if !info.IsDir() {
		return nil, result.Validation(result.CodeInvalidInput, "fs.list requires a directory", map[string]any{"path": rel})
	}

	var entries []listEntry
	walkErr := filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return result.Runtime(result.CodeCancelled, "filesystem cancelled", false, nil)
		}
		if p == abs {
			if !in.Recursive {
				return nil
			}
			return nil
		}
		wsRoot, werr := workspaceRoot(req.Workspace)
		if werr != nil {
			return werr
		}
		relEntry, err := filepath.Rel(wsRoot, p)
		if err != nil {
			relEntry, _ = filepath.Rel(abs, p)
		}
		relEntry = filepath.ToSlash(relEntry)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := rejectEscapingSymlink(req.Workspace, p, info); err != nil {
			return err
		}
		entries = append(entries, listEntry{Path: relEntry, Type: nodeType(info)})
		if !in.Recursive && d.IsDir() {
			return fs.SkipDir
		}
		if in.Recursive && d.Type()&os.ModeSymlink != 0 {
			return fs.SkipDir
		}
		return nil
	})
	if walkErr != nil {
		if re, ok := walkErr.(result.Error); ok {
			return nil, re
		}
		return nil, result.Runtime(result.CodePlayerError, "fs.list: "+walkErr.Error(), false, nil)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	truncated := len(entries) > limit
	if truncated {
		entries = entries[:limit]
	}
	if entries == nil {
		entries = []listEntry{}
	}
	return json.Marshal(listOut{Path: rel, Entries: entries, Truncated: truncated})
}

func execStat(req registry.ExecRequest) (json.RawMessage, error) {
	var in statIn
	_ = json.Unmarshal(req.Input, &in)
	abs, rel, err := resolvePath(req.Workspace, in.Path, resolveOpts{allowRoot: true})
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, result.Validation(result.CodeInvalidInput, "fs.stat path not found", map[string]any{"path": rel})
	}
	if err := rejectEscapingSymlink(req.Workspace, abs, info); err != nil {
		return nil, err
	}
	return json.Marshal(statOut{
		Path:       rel,
		Type:       nodeType(info),
		Size:       info.Size(),
		Mode:       fmt.Sprintf("%04o", info.Mode().Perm()),
		ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
	})
}

func validateWriteContent(content string) error {
	if !utf8.ValidString(content) {
		return result.Validation(result.CodeInvalidInput, "fs.write content must be valid UTF-8", nil)
	}
	if len(content) > maxBytesLimit {
		return result.Validation(result.CodeInvalidInput, "fs.write content exceeds 4MiB limit", map[string]any{
			"bytes": len(content),
			"max":   maxBytesLimit,
		})
	}
	return nil
}

func clampBytes(n int) (int, error) {
	if n == 0 {
		return defaultMaxBytes, nil
	}
	if n < 1 || n > maxBytesLimit {
		return 0, result.Validation(result.CodeInvalidInput, "max_bytes must be between 1 and 4194304", map[string]any{"max_bytes": n})
	}
	return n, nil
}

func clampEntries(n int) (int, error) {
	if n == 0 {
		return defaultMaxEntries, nil
	}
	if n < 1 || n > maxEntriesLimit {
		return 0, result.Validation(result.CodeInvalidInput, "max_entries must be between 1 and 1000", map[string]any{"max_entries": n})
	}
	return n, nil
}

func nodeType(info os.FileInfo) string {
	if info.Mode()&os.ModeSymlink != 0 {
		return "symlink"
	}
	if info.IsDir() {
		return "directory"
	}
	return "file"
}
