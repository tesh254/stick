package crawl

import (
	"net/http"
	"sync"
	"time"
)

type CrawlConfig struct {
	UserAgent    string
	Timeout      time.Duration
	RequestDelay time.Duration
}

type Metadata struct {
}

type LlmTxtContent struct {
	URL             string
	Content         *string
	Config          *CrawlConfig
	client          *http.Client
	lastRequestTime map[string]time.Time
	mutex           sync.Mutex
}
