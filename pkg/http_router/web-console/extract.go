package web

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"go.uber.org/zap"
)

const (
	extractedDirName = "web_console"
	stampFileName    = ".ui-stamp"
	indexFileName    = "index.html"
)

// Ensure prepares cfg.Web.WebDirectory.
//
// If the user already set web_directory, that path is left unchanged.
// Otherwise, when this binary contains a packed UI, the archive is extracted
// into {data}/internal/web_console (skipped when the stamp still matches)
// and web_directory is pointed at that folder. The HTTP handler then serves
// files from disk, so the embedded bytes are not kept in the heap.
func Ensure(logger *zap.Logger, cfg *config.Config) error {
	return ensure(logger, cfg, embeddedZIP)
}

// HasEmbeddedUI reports whether the binary contains a packed console (index.html).
func HasEmbeddedUI() bool {
	return zipContainsIndex(embeddedZIP)
}

func ensure(logger *zap.Logger, cfg *config.Config, zipData []byte) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	if strings.TrimSpace(cfg.Web.WebDirectory) != "" {
		logger.Info("using configured web console directory", zap.String("web_directory", cfg.Web.WebDirectory))
		return nil
	}

	if !zipContainsIndex(zipData) {
		logger.Info("no embedded web console in this binary; set web.web_directory or pack the UI at build time")
		return nil
	}

	dest := filepath.Join(cfg.Directories.GetDataInternal(), extractedDirName)
	stamp := uiStamp(zipData)
	if sameStamp(dest, stamp) && fileExists(filepath.Join(dest, indexFileName)) {
		cfg.Web.WebDirectory = dest
		logger.Info("using extracted web console", zap.String("web_directory", dest))
		return nil
	}

	start := time.Now()
	if err := extractTo(zipData, dest, stamp); err != nil {
		return fmt.Errorf("extract web console: %w", err)
	}
	cfg.Web.WebDirectory = dest
	logger.Info("extracted embedded web console",
		zap.String("destination", dest),
		zap.String("stamp", stamp),
		zap.String("timeTaken", time.Since(start).String()),
	)
	return nil
}

func uiStamp(zipData []byte) string {
	sum := sha256.Sum256(zipData)
	v := version.Get()
	return fmt.Sprintf("%s+%s+%s", v.Version, v.GitCommit, hex.EncodeToString(sum[:8]))
}

func sameStamp(dest, stamp string) bool {
	b, err := os.ReadFile(filepath.Join(dest, stampFileName))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b)) == stamp
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func zipContainsIndex(zipData []byte) bool {
	if len(zipData) == 0 {
		return false
	}
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return false
	}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(filepath.ToSlash(file.Name)) == indexFileName {
			return true
		}
	}
	return false
}

func extractTo(zipData []byte, dest, stamp string) error {
	tmp := dest + ".extracting"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}

	if err := unzipBytes(zipData, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if !fileExists(filepath.Join(tmp, indexFileName)) {
		_ = os.RemoveAll(tmp)
		return fmt.Errorf("packed UI is missing %s", indexFileName)
	}
	if err := os.WriteFile(filepath.Join(tmp, stampFileName), []byte(stamp+"\n"), 0o644); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}

	if err := os.RemoveAll(dest); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	return nil
}

func unzipBytes(zipData []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	dest = filepath.Clean(dest)
	destPrefix := dest + string(os.PathSeparator)

	for _, file := range reader.File {
		name := filepath.ToSlash(file.Name)
		if name == "" || strings.HasPrefix(name, "__MACOSX/") || strings.HasSuffix(strings.ToLower(name), ".map") {
			continue
		}

		target := filepath.Join(dest, filepath.FromSlash(name))
		if target != dest && !strings.HasPrefix(target, destPrefix) {
			return fmt.Errorf("%s is an illegal filepath", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		if err := writeZipFile(file, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(file *zip.File, target string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
