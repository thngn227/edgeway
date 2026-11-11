package configmodel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadStatic(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid with defaults",
			yaml: "entryPoints:\n  web:\n    address: \":8000\"\n",
		},
		{
			name:    "no entry points",
			yaml:    "log:\n  level: info\n",
			wantErr: "at least one entry point",
		},
		{
			name:    "unknown key is rejected",
			yaml:    "entryPoints:\n  web:\n    adress: \":8000\"\n",
			wantErr: "field adress not found",
		},
		{
			name:    "bad log level",
			yaml:    "log:\n  level: verbose\nentryPoints:\n  web:\n    address: \":8000\"\n",
			wantErr: "log.level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := LoadStatic(writeFile(t, tt.yaml))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := s.EntryPoints["web"].Timeouts.ReadHeader; got != 10*time.Second {
				t.Errorf("default readHeader timeout = %v, want 10s", got)
			}
			if s.LifeCycle.GraceTimeout != 10*time.Second {
				t.Errorf("default graceTimeout = %v, want 10s", s.LifeCycle.GraceTimeout)
			}
		})
	}
}

func TestLoadDynamic(t *testing.T) {
	entryPoints := map[string]EntryPoint{"web": {Address: ":8000"}}

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid",
			yaml: `
routers:
  api:
    entryPoints: [web]
    rule: "PathPrefix(` + "`/`" + `)"
    service: api
services:
  api:
    loadBalancer:
      servers:
        - url: http://localhost:9001
`,
		},
		{
			name: "unknown service",
			yaml: `
routers:
  api:
    rule: "PathPrefix(` + "`/`" + `)"
    service: missing
`,
			wantErr: `unknown service "missing"`,
		},
		{
			name: "unknown entry point",
			yaml: `
routers:
  api:
    entryPoints: [websecure]
    rule: "PathPrefix(` + "`/`" + `)"
    service: api
services:
  api:
    loadBalancer:
      servers:
        - url: http://localhost:9001
`,
			wantErr: `unknown entry point "websecure"`,
		},
		{
			name: "server url without scheme",
			yaml: `
services:
  api:
    loadBalancer:
      servers:
        - url: localhost:9001
`,
			wantErr: "scheme must be http or https",
		},
		{
			name: "service without servers",
			yaml: `
services:
  api:
    loadBalancer:
      servers: []
`,
			wantErr: "at least one server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadDynamic(writeFile(t, tt.yaml), entryPoints)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
