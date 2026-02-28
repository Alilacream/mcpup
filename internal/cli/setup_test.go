package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mohammedsamin/mcpup/internal/store"
)

func TestSetupRequiresServerInNonInteractiveMode(t *testing.T) {
	env := setupTestEnv(t)
	runCLI(t, env, "init")

	var stderr bytes.Buffer
	err := Run([]string{"setup", "--client", "cursor"}, nil, &bytes.Buffer{}, &stderr)
	if err == nil {
		t.Fatalf("expected setup without --server to fail in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "setup requires at least one --server") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupConfiguresServerAndEnablesClient(t *testing.T) {
	env := setupTestEnv(t)
	runCLI(t, env, "init")

	runCLI(t, env,
		"setup",
		"--client", "cursor",
		"--server", "github",
		"--env", "GITHUB_TOKEN=test-token",
		"--yes",
	)

	cfg, err := store.LoadConfig(env.configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	srv, ok := cfg.Servers["github"]
	if !ok {
		t.Fatalf("expected github server in canonical config")
	}
	if srv.Command != "npx" {
		t.Fatalf("expected registry command npx, got %q", srv.Command)
	}
	if srv.Env["GITHUB_TOKEN"] != "test-token" {
		t.Fatalf("expected env var to be stored from setup")
	}
	if !cfg.Clients["cursor"].Servers["github"].Enabled {
		t.Fatalf("expected github enabled on cursor client")
	}

	cursorPath := filepath.Join(env.home, ".cursor", "mcp.json")
	body, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatalf("read cursor config: %v", err)
	}
	doc := map[string]map[string]map[string]any{}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("parse cursor config: %v", err)
	}
	enabled, _ := doc["mcpServers"]["github"]["enabled"].(bool)
	if !enabled {
		t.Fatalf("expected github enabled=true in cursor config")
	}
}
