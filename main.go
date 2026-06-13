package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const InitMessage = "Docker Swarm Utils: initialized"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "entrypoint" {
		runEntrypoint()
		return
	}
	fmt.Println(InitMessage)
}

func runEntrypoint() {
	// Check if command is provided
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: docker-swarm-utils entrypoint <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[2]
	args := os.Args[2:]

	// Prepare environment variables
	// TODO: extend environment variables with those from Docker Secrets
	env := append(os.Environ(), "FOO=BAR")

	// Prepare the command
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Replace the current process with the specified command
	err = syscall.Exec(path, args, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
		os.Exit(1)
	}
}
