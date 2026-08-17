package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/result"
)

type resolveOpts struct {
	allowRoot     bool
	allowMissing  bool
	forWrite      bool
	createParents bool
}

func workspaceRoot(workspace string) (string, error) {
	ws, err := filepath.Abs(workspace)
	if err != nil {
		return "", result.Runtime(result.CodeInternal, err.Error(), false, nil)
	}
	if resolved, err := filepath.EvalSymlinks(ws); err == nil {
		ws = resolved
	}
	return ws, nil
}

func resolvePath(workspace, rel string, opts resolveOpts) (abs string, relOut string, err error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		rel = "."
	}
	if filepath.IsAbs(rel) {
		return "", "", result.Validation(result.CodeInvalidInput, "path must be relative to workspace_root", map[string]any{"path": rel})
	}
	cleaned := filepath.Clean(rel)
	if filepath.IsAbs(cleaned) {
		return "", "", result.Validation(result.CodeInvalidInput, "path must be relative to workspace_root", map[string]any{"path": rel})
	}

	ws, err := workspaceRoot(workspace)
	if err != nil {
		return "", "", err
	}

	var resolved string
	if cleaned == "." {
		resolved = ws
	} else {
		parentRel := filepath.Dir(cleaned)
		leaf := filepath.Base(cleaned)
		parentAbs := ws
		if parentRel != "." {
			parentAbs, err = evalExisting(filepath.Join(ws, parentRel))
			if err != nil {
				if !opts.createParents || !os.IsNotExist(err) {
					return "", "", result.Validation(result.CodeInvalidInput, "invalid path: "+err.Error(), map[string]any{"path": rel})
				}
				parentAbs = filepath.Join(ws, parentRel)
			}
			if err := confined(ws, parentAbs); err != nil {
				return "", "", err
			}
		}
		resolved = filepath.Join(parentAbs, leaf)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", "", result.Validation(result.CodeInvalidInput, "invalid path", map[string]any{"path": rel})
	}
	if err := confined(ws, resolved); err != nil {
		return "", "", err
	}

	display, err := filepath.Rel(ws, resolved)
	if err != nil {
		display = cleaned
	}
	display = filepath.ToSlash(display)
	if display == "." {
		if opts.forWrite || !opts.allowRoot {
			return "", "", result.Validation(result.CodeInvalidInput, "path must not be the workspace root", map[string]any{"path": rel})
		}
	}

	if opts.forWrite {
		if err := validateWriteDest(ws, resolved, opts); err != nil {
			return "", "", err
		}
	}
	return resolved, display, nil
}

func validateWriteDest(ws, resolved string, opts resolveOpts) error {
	parent := filepath.Dir(resolved)
	if err := confined(ws, parent); err != nil {
		return err
	}
	info, err := os.Lstat(resolved)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return result.Validation(result.CodeInvalidInput, "fs.write destination must not be a symlink", map[string]any{"path": resolved})
		}
		if info.IsDir() {
			return result.Validation(result.CodeInvalidInput, "fs.write destination must not be a directory", map[string]any{"path": resolved})
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return result.Runtime(result.CodePlayerError, "fs.write stat: "+err.Error(), false, nil)
	}
	_, err = os.Lstat(parent)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		if !opts.createParents {
			return result.Validation(result.CodeInvalidInput, "fs.write parent directory does not exist", map[string]any{"path": resolved})
		}
		return confined(ws, parent)
	}
	return result.Runtime(result.CodePlayerError, "fs.write parent: "+err.Error(), false, nil)
}

func confined(ws, path string) error {
	rel, err := filepath.Rel(ws, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return result.Validation(result.CodeInvalidInput, "path must be inside workspace_root", map[string]any{
			"workspace": ws,
			"path":      path,
		})
	}
	return nil
}

func evalExisting(path string) (string, error) {
	if r, err := filepath.EvalSymlinks(path); err == nil {
		return r, nil
	}
	var missing []string
	p := path
	for {
		dir := filepath.Dir(p)
		base := filepath.Base(p)
		if dir == p {
			return "", os.ErrNotExist
		}
		missing = append([]string{base}, missing...)
		if r, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(append([]string{r}, missing...)...), nil
		}
		p = dir
	}
}

func rejectEscapingSymlink(workspace, path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	_, err := confinedSymlinkTarget(workspace, path)
	return err
}

func confinedSymlinkTarget(workspace, link string) (string, error) {
	ws, err := workspaceRoot(workspace)
	if err != nil {
		return "", err
	}
	raw, err := os.Readlink(link)
	if err != nil {
		return "", result.Validation(result.CodeInvalidInput, "invalid symlink: "+err.Error(), map[string]any{"path": link})
	}
	target := raw
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), raw)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", result.Validation(result.CodeInvalidInput, "invalid symlink target", map[string]any{"path": link})
	}
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		target = resolved
	}
	if err := confined(ws, target); err != nil {
		return "", result.Validation(result.CodeInvalidInput, "symlink target must be inside workspace_root", map[string]any{
			"path":   link,
			"target": target,
		})
	}
	return target, nil
}
