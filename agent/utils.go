package agent

import "strings"

// CountTokens provides a rough estimate of the number of tokens in the messages.
func (c *Client) CountTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += len(strings.Split(msg.Content, " "))
	}
	return total
}
