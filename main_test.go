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

func TestGetVariablesFromDir_DotEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dotenv-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Case 1: Standard secret
	os.WriteFile(filepath.Join(tempDir, "DB_PASS"), []byte("secret123"), 0644)

	// Case 2: Dotenv secret with hint
	dotenvContent := "# format: dotenv\nFOO=bar\nBAZ=qux\n"
	os.WriteFile(filepath.Join(tempDir, "app.env"), []byte(dotenvContent), 0644)

	// Case 3: Override - lexical order
	os.WriteFile(filepath.Join(tempDir, "00_base"), []byte("# format: dotenv\nCOMMON=base\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "01_override"), []byte("# format: dotenv\nCOMMON=override\n"), 0644)

	got, err := getVariablesFromDir(tempDir)
	if err != nil {
		t.Errorf("getVariablesFromDir() error = %v", err)
	}

	expected := map[string]string{
		"DB_PASS": "secret123",
		"FOO":     "bar",
		"BAZ":     "qux",
		"COMMON":  "override",
	}

	for k, v := range expected {
		if val, ok := got[k]; !ok || val != v {
			t.Errorf("expected %s=%s, got %s", k, v, val)
		}
	}

	if len(got) != 4 {
		t.Errorf("expected 4 variables, got %d: %v", len(got), got)
	}
}

func TestGetVariablesFromDir_InvalidDotEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invalid-dotenv-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// File with hint but invalid syntax (godotenv might be lenient, but let's try something it should fail on if possible)
	// Actually godotenv.Unmarshal is quite lenient. 
	// Let's see if we can trigger an error. 
	// According to godotenv source, it might fail if there's no '=' in a line that isn't a comment or empty, 
	// BUT only if we use Parse. Unmarshal might behave differently.
	
	// Let's try to use a case that SHOULD be invalid.
	content := "# format: dotenv\nINVALID_LINE_WITHOUT_EQUALS\n"
	os.WriteFile(filepath.Join(tempDir, "invalid.env"), []byte(content), 0644)
	os.WriteFile(filepath.Join(tempDir, "VALID"), []byte("value"), 0644)

	got, err := getVariablesFromDir(tempDir)
	if err != nil {
		t.Errorf("getVariablesFromDir() error = %v", err)
	}

	// If godotenv.Unmarshal fails, it should skip the file.
	// If it doesn't fail (because it's lenient), it might just not find variables.
	
	if got["VALID"] != "value" {
		t.Errorf("expected VALID=value, got %s", got["VALID"])
	}
}

func TestIsValidEnvKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"VALID", true},
		{"valid_name_123", true},
		{"_START_WITH_UNDERSCORE", true},
		{"123_INVALID_START", false},
		{"INVALID-CHAR", false},
		{"INVALID SPACE", false},
		{"VAR;echo injection", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isValidEnvKey(tt.key); got != tt.want {
			t.Errorf("isValidEnvKey(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestLoadVariables_SkipsInvalidKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invalid-keys-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	os.WriteFile(filepath.Join(tempDir, "VALID"), []byte("value"), 0644)
	os.WriteFile(filepath.Join(tempDir, "INVALID;KEY"), []byte("value"), 0644)

	var buf bytes.Buffer
	err = loadVariables(&buf, tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("export VALID='value'")) {
		t.Error("Expected output to contain VALID")
	}
	if bytes.Contains(buf.Bytes(), []byte("INVALID;KEY")) {
		t.Error("Expected output NOT to contain INVALID;KEY")
	}
}
