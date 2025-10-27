package prompts

func NewPrompts() Prompts {
	return Prompts{
		SystemPrompt:   systemPrompt,
		TitleGenerator: titleGenerator,
	}
}
