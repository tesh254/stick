package handlers

import (
	"fmt"

	"github.com/tesh254/stick/internal/tui"
)

func StartSession() {
	p := tui.NewProgram()
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
