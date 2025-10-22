package utils

import (
	"os"
	"os/user"
)

func GetUser() string {
	u, _ := user.Current()

	return u.Username
}

func GetHomeDir() string {
	home, _ := os.UserHomeDir()

	return home
}
