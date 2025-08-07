package agent

// AIClient is the interface for AI clients.
type AIClient interface {
	Create(req ChatCompletionRequest) (*ChatCompletionResponse, error)
	MessageStream(req ChatCompletionRequest) (chan string, error)
}
