package provider

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// toolPayload mirrors the output payload for easy testing
type toolPayload struct {
    Tool     string            `json:"tool"`
    Status   string            `json:"status"`
    Data     string            `json:"data"`
    Metadata map[string]string `json:"metadata"`
}

func readToolPayload(t *testing.T, out string, ch <-chan string) toolPayload {
    t.Helper()
    // Read from channel (actual payload)
    payloadLine := <-ch
    if !strings.HasPrefix(payloadLine, "[TOOL_OUTPUT] ") {
        t.Fatalf("unexpected payload prefix: %s", payloadLine)
    }
    jsonStr := strings.TrimPrefix(payloadLine, "[TOOL_OUTPUT] ")
    var p toolPayload
    if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
        t.Fatalf("unmarshal: %v\nraw: %s", err, jsonStr)
    }
    return p
}

func TestGetPageContent_Success(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        io := `<html><head><title>Test Page</title></head><body><h1>Hello World</h1><p>This is a test.</p></body></html>`
        _, _ = w.Write([]byte(io))
    }))
    defer srv.Close()

    final, ch := getPageContent(GetPageContentArgs{URL: srv.URL})
    _ = final // ignore immediate placeholder
    p := readToolPayload(t, final, ch)
    if p.Status != "success" {
        t.Fatalf("expected success, got %s: %s", p.Status, p.Data)
    }
    if !strings.Contains(p.Data, "Hello World") {
        t.Fatalf("markdown should contain body content, got: %s", p.Data)
    }
}

func TestGetPageContent_InvalidURL(t *testing.T) {
    final, ch := getPageContent(GetPageContentArgs{URL: "not-a-url"})
    _ = final
    p := readToolPayload(t, final, ch)
    if p.Status != "error" {
        t.Fatalf("expected error for invalid URL, got %s", p.Status)
    }
}

func TestWebsearch_ParsesResults(t *testing.T) {
    // Mock search page
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        // Two results with anchor and h3
        html := `
        <html><body>
          <div class="result"><a href="http://example.com/one"><h3>Result One</h3></a><div>First snippet content.</div></div>
          <div class="result"><a href="https://example.org/two"><h3>Result Two</h3></a><div>Second snippet content.</div></div>
        </body></html>`
        _, _ = w.Write([]byte(html))
    }))
    defer srv.Close()

    t.Setenv("STICK_SEARCH_BASE_URL", srv.URL)
    final, ch := websearch(WebsearchArgs{Prompt: "testing", TopK: 2})
    _ = final
    p := readToolPayload(t, final, ch)
    if p.Status != "success" {
        t.Fatalf("expected success, got %s: %s", p.Status, p.Data)
    }
    if !strings.Contains(p.Data, "Result One") || !strings.Contains(p.Data, "Result Two") {
        t.Fatalf("expected structured headings for results, got: %s", p.Data)
    }
    if p.Metadata["count"] != "2" {
        t.Fatalf("expected count=2, got %s", p.Metadata["count"])
    }
}

// Benchmarks
func BenchmarkGetPageContentLarge(b *testing.B) {
    largeHTML := strings.Repeat("<div><p>Lorem ipsum dolor sit amet.</p></div>", 5000)
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        _, _ = w.Write([]byte("<html><body>" + largeHTML + "</body></html>"))
    }))
    defer srv.Close()

    for i := 0; i < b.N; i++ {
        _, ch := getPageContent(GetPageContentArgs{URL: srv.URL})
        <-ch
    }
}

func BenchmarkWebsearchLarge(b *testing.B) {
    largeHTML := strings.Repeat("<div class='result'><a href='http://example.com/x'><h3>Title</h3></a><div>Snippet</div></div>", 5000)
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        _, _ = w.Write([]byte("<html><body>" + largeHTML + "</body></html>"))
    }))
    defer srv.Close()

    b.ResetTimer()
    tTop := 100
    for i := 0; i < b.N; i++ {
        // limit top_k to prevent huge markdown
        _, ch := websearch(WebsearchArgs{Prompt: "bench", TopK: tTop})
        <-ch
    }
}