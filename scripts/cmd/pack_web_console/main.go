package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}

	src := filepath.Join(root, "web-console", "build")
	indexPath := filepath.Join(src, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		fatal(fmt.Errorf("web-console/build/index.html not found; run ./scripts/build_web_console.sh first: %w", err))
	}
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		fatal(err)
	}
	if bytes.Contains(indexHTML, []byte("/src/index.js")) {
		fatal(fmt.Errorf("web-console/build/index.html is the Vite source page, not a production build"))
	}

	destDir := filepath.Join(root, "pkg", "http_router", "web-console", "assets")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		fatal(err)
	}
	dest := filepath.Join(destDir, "web_console.zip")

	tmp, err := os.CreateTemp(destDir, "web_console-*.zip")
	if err != nil {
		fatal(err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	zw := zip.NewWriter(tmp)
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if skipPackedName(rel, d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			_, err := zw.Create(rel + "/")
			return err
		}
		return writeFile(zw, path, rel)
	})
	if err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		fatal(err)
	}
	if err := zw.Close(); err != nil {
		_ = tmp.Close()
		fatal(err)
	}
	if err := tmp.Close(); err != nil {
		fatal(err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		fatal(err)
	}

	info, err := os.Stat(dest)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("packed %s (%d bytes)\n", dest, info.Size())
}

func writeFile(zw *zip.Writer, absPath, rel string) error {
	w, err := zw.Create(rel)
	if err != nil {
		return err
	}
	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(w, f)
	return err
}

func skipPackedName(rel string, isDir bool) bool {
	base := filepath.Base(rel)
	if base == ".DS_Store" || strings.HasPrefix(base, ".") {
		return true
	}
	if !isDir && strings.HasSuffix(strings.ToLower(rel), ".map") {
		return true
	}
	return false
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "pack_web_console: %v\n", err)
	os.Exit(1)
}
