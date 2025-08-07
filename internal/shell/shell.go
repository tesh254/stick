package shell

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func getDefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	if currentUser, err := user.Current(); err == nil {
		file, err := os.Open("/etc/passwd")
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fields := strings.Split(scanner.Text(), ":")
				if len(fields) >= 7 && fields[0] == currentUser.Username {
					if shell := fields[6]; shell != "" {
						return shell
					}
				}
			}
		}
	}

	return "/bin/sh"
}

func ExecuteCommand(command string) *exec.Cmd {
	shell := getDefaultShell()
	shellName := filepath.Base(shell)

	var cmd *exec.Cmd
	switch shellName {
	case "bash":
		cmd = exec.Command(shell, "--login", "-c", command)
	case "zsh":
		cmd = exec.Command(shell, "-l", "-c", command)
	case "fish":
		cmd = exec.Command(shell, "-l", "-c", command)
	default:
		cmd = exec.Command(shell, "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd
}
