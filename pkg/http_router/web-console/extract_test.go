package web

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"go.uber.org/zap"
)

func TestEnsureUsesConfiguredDirectory(t *testing.T) {
	cfg := &config.Config{
		Web:         config.WebConfig{WebDirectory: "/custom/ui"},
		Directories: config.Directories{Data: t.TempDir()},
	}
	if err := ensure(zap.NewNop(), cfg, mustZip(t, map[string]string{indexFileName: "<html/>"})); err != nil {
		t.Fatal(err)
	}
	if cfg.Web.WebDirectory != "/custom/ui" {
		t.Fatalf("web_directory = %q, want configured path", cfg.Web.WebDirectory)
	}
}

func TestEnsureSkipsEmptyArchive(t *testing.T) {
	cfg := &config.Config{
		Directories: config.Directories{Data: t.TempDir()},
	}
	if err := ensure(zap.NewNop(), cfg, mustZip(t, map[string]string{"static/app.js": "1"})); err != nil {
		t.Fatal(err)
	}
	if cfg.Web.WebDirectory != "" {
		t.Fatalf("web_directory = %q, want empty", cfg.Web.WebDirectory)
	}
}

func TestEnsureExtractsAndReusesStamp(t *testing.T) {
	dataDir := t.TempDir()
	zipData := mustZip(t, map[string]string{
		indexFileName:          "<html>ok</html>",
		"static/app.js":        "console.log(1)",
		"static/app.js.map":    "SHOULD_NOT_BE_EXTRACTED",
		"locales/en/file.yaml": "ok: true",
	})

	cfg := &config.Config{Directories: config.Directories{Data: dataDir}}
	if err := ensure(zap.NewNop(), cfg, zipData); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dataDir, "internal", extractedDirName)
	if cfg.Web.WebDirectory != dest {
		t.Fatalf("web_directory = %q, want %q", cfg.Web.WebDirectory, dest)
	}
	if got := readFile(t, filepath.Join(dest, indexFileName)); got != "<html>ok</html>" {
		t.Fatalf("index.html = %q", got)
	}
	if fileExists(filepath.Join(dest, "static", "app.js.map")) {
		t.Fatal("source map should not be extracted")
	}

	// second start with the same zip must not rewrite the tree
	indexPath := filepath.Join(dest, indexFileName)
	info1, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Web.WebDirectory = ""
	if err := ensure(zap.NewNop(), cfg, zipData); err != nil {
		t.Fatal(err)
	}
	info2, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Fatal("expected stamp match to skip extract")
	}
}

func TestEnsureReextractsWhenArchiveChanges(t *testing.T) {
	dataDir := t.TempDir()
	cfg := &config.Config{Directories: config.Directories{Data: dataDir}}

	if err := ensure(zap.NewNop(), cfg, mustZip(t, map[string]string{indexFileName: "v1"})); err != nil {
		t.Fatal(err)
	}
	dest := cfg.Web.WebDirectory
	if got := readFile(t, filepath.Join(dest, indexFileName)); got != "v1" {
		t.Fatalf("first extract = %q", got)
	}

	cfg.Web.WebDirectory = ""
	if err := ensure(zap.NewNop(), cfg, mustZip(t, map[string]string{indexFileName: "v2"})); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dest, indexFileName)); got != "v2" {
		t.Fatalf("second extract = %q, want v2", got)
	}
}

func TestUnzipRejectsPathTraversal(t *testing.T) {
	dest := t.TempDir()
	err := unzipBytes(mustZip(t, map[string]string{"../outside.txt": "nope"}), dest)
	if err == nil {
		t.Fatal("expected path traversal to fail")
	}
}

func TestZipContainsIndex(t *testing.T) {
	if zipContainsIndex(nil) || zipContainsIndex([]byte("not-a-zip")) {
		t.Fatal("invalid data should not report an index")
	}
	if zipContainsIndex(mustZip(t, map[string]string{"static/app.js": "1"})) {
		t.Fatal("archive without index.html should be empty")
	}
	if !zipContainsIndex(mustZip(t, map[string]string{indexFileName: "<html/>"})) {
		t.Fatal("archive with index.html should be detected")
	}
}

func mustZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
