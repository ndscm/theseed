package common_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ndscm/theseed/seed/devprod/go/common"
)

func TestDownloadFile(t *testing.T) {
	tmpDir := os.Getenv("TEST_TMPDIR")
	if tmpDir == "" {
		t.Fatal("TEST_TMPDIR is not set")
	}

	targetPath := filepath.Join(tmpDir, "baidu.html")
	err := common.DownloadFile("https://www.baidu.com", targetPath)
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}
	bodyBytes, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read body bytes: %v", err)
	}
	if !strings.Contains(string(bodyBytes), "login") {
		t.Fatalf("Failed to download: %v", err)
	}
}
