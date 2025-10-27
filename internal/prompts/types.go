package prompts

type Prompts struct {
	SystemPrompt   string
	TitleGenerator func(string) string
}
