package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitMessage(t *testing.T) {
	expected := "Docker Swarm Utils: initialized"
	if InitMessage != expected {
		t.Errorf("expected %q, got %q", expected, InitMessage)
	}
}

func TestParseEntrypointArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCmd     string
		wantArgLen  int
		wantErr     bool
	}{
		{
			name:       "valid command",
			args:       []string{"docker-swarm-utils", "entrypoint", "python", "app.py"},
			wantCmd:    "python",
			wantArgLen: 2,
			wantErr:    false,
		},
		{
			name:       "missing command",
			args:       []string{"docker-swarm-utils", "entrypoint"},
			wantCmd:    "",
			wantArgLen: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := parseEntrypointArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEntrypointArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cmd != tt.wantCmd {
				t.Errorf("parseEntrypointArgs() cmd = %v, want %v", cmd, tt.wantCmd)
			}
			if len(args) != tt.wantArgLen {
				t.Errorf("parseEntrypointArgs() args len = %v, want %v", len(args), tt.wantArgLen)
			}
		})
	}
}

func TestGetVariablesFromDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "secrets-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	secrets := map[string]string{
		"DB_PASS": "secret123\n",
		"API_KEY": "abc-xyz",
	}

	for k, v := range secrets {
		err := os.WriteFile(filepath.Join(tempDir, k), []byte(v), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	got, err := getVariablesFromDir(tempDir)
	if err != nil {
		t.Errorf("getVariablesFromDir() error = %v", err)
	}

	if len(got) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(got))
	}

	if got["DB_PASS"] != "secret123" {
		t.Errorf("expected DB_PASS=secret123, got %s", got["DB_PASS"])
	}
}

func TestLoadVariables(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "load-vars-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.WriteFile(filepath.Join(tempDir, "TEST_VAR"), []byte("value-with-'quote'"), 0644)

	var buf bytes.Buffer
	err = loadVariables(&buf, tempDir)
	if err != nil {
		t.Errorf("loadVariables() error = %v", err)
	}

	got := buf.String()
	expected := "export TEST_VAR='value-with-'\\''quote'\\'''\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestExtendEnvWithDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "env-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.WriteFile(filepath.Join(tempDir, "MY_VAR"), []byte("my-value"), 0644)

	baseEnv := []string{"FOO=BAR"}
	got := extendEnvWithDir(baseEnv, tempDir)

	// Test second call to ensure it appends
	tempDir2, err := os.MkdirTemp("", "env-test-2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir2)
	os.WriteFile(filepath.Join(tempDir2, "OTHER_VAR"), []byte("other-value"), 0644)

	got = extendEnvWithDir(got, tempDir2)

	foundFoo := false
	foundMyVar := false
	foundOtherVar := false

	for _, e := range got {
		if e == "FOO=BAR" {
			foundFoo = true
		}
		if e == "MY_VAR=my-value" {
			foundMyVar = true
		}
		if e == "OTHER_VAR=other-value" {
			foundOtherVar = true
		}
	}

	if !foundFoo {
		t.Error("FOO=BAR missing from extended env")
	}
	if !foundMyVar {
		t.Error("MY_VAR=my-value missing from extended env")
	}
	if !foundOtherVar {
		t.Error("OTHER_VAR=other-value missing from extended env")
	}
}
