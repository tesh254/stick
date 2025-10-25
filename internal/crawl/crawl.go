package crawl

import (
	"net/http"
	"time"
)

func DefaultConfig() *CrawlConfig {
	return &CrawlConfig{
		UserAgent:    "Mozilla/5.0 (compatible; StickScraper/1.0)",
		Timeout:      60 * time.Second,
		RequestDelay: 3 * time.Second,
	}
}

func NewLlmText(url string, config *CrawlConfig) *LlmTxtContent {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	l := &LlmTxtContent{
		URL:             url,
		Config:          config,
		client:          client,
		lastRequestTime: make(map[string]time.Time),
	}

	return l
}

func (l *LlmTxtContent) GetContent(url string) (string, error) {
	return "", nil
}

func (c *CrawlConfig) makeFetchCall() {

}
