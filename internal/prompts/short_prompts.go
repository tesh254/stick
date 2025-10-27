package prompts

import "fmt"

func titleGenerator(prompt string) string {
	return fmt.Sprintf(`Generate short title based on this prompt by the user '%s'
		call tool title_generator with the title as the argument
		`, prompt)
}
