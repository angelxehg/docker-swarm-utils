package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

const InitMessage = "Docker Swarm Utils: initialized"

func main() {
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "entrypoint":
			runEntrypoint()
			return
		case "load-variables", "load-variables-from-dir":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: %s %s <directory>\n", os.Args[0], cmd)
				os.Exit(1)
			}
			err := loadVariables(os.Stdout, os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Println(InitMessage)
}

// loadVariables prints shell export commands for variables found in the directory.
func loadVariables(w io.Writer, dirPath string) error {
	variables, err := getVariablesFromDir(dirPath)
	if err != nil {
		return err
	}

	for key, value := range variables {
		if !isValidEnvKey(key) {
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid environment variable name: %s\n", key)
			continue
		}
		// Escape single quotes for shell safety: ' -> '\''
		escapedValue := strings.ReplaceAll(value, "'", "'\\''")
		fmt.Fprintf(w, "export %s='%s'\n", key, escapedValue)
	}

	return nil
}

func runEntrypoint() {
	command, args, err := parseEntrypointArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// TODO: make paths configurable via a file.
	env := os.Environ()
	for _, dir := range []string{"/run/secrets", "/run/configs"} {
		env = extendEnvWithDir(env, dir)
	}

	err = executeCommand(command, args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
		os.Exit(1)
	}
}

func parseEntrypointArgs(args []string) (string, []string, error) {
	if len(args) < 3 {
		return "", nil, fmt.Errorf("Usage: docker-swarm-utils entrypoint <command> [args...]")
	}
	return args[2], args[2:], nil
}

func extendEnvWithDir(baseEnv []string, dirPath string) []string {
	env := baseEnv

	variables, err := getVariablesFromDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load variables from %s: %v\n", dirPath, err)
		return env
	}

	for k, v := range variables {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func getVariablesFromDir(dirPath string) (map[string]string, error) {
	variables := make(map[string]string)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return variables, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read file %s: %v\n", entry.Name(), err)
			continue
		}

		strContent := string(content)
		if strings.HasPrefix(strContent, "# format: dotenv") {
			parsed, err := godotenv.Unmarshal(strContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse dotenv file %s: %v\n", entry.Name(), err)
				continue
			}
			for k, v := range parsed {
				variables[k] = v
			}
		} else {
			variables[entry.Name()] = strings.TrimSpace(strContent)
		}
	}

	return variables, nil
}

func executeCommand(command string, args []string, env []string) error {
	path, err := exec.LookPath(command)
	if err != nil {
		return err
	}

	return syscall.Exec(path, args, env)
}

func isValidEnvKey(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}
