package fbevents

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultHeaders mimics a Chrome browser request to pass Facebook's bot checks.
var defaultHeaders = map[string]string{
	"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
"accept-language":           "en-US,en;q=0.6",
	"cache-control":             "max-age=0",
	"sec-fetch-dest":            "document",
	"sec-fetch-mode":            "navigate",
	"sec-fetch-site":            "same-origin",
	"sec-fetch-user":            "?1",
	"sec-gpc":                   "1",
	"upgrade-insecure-requests": "1",
	"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
}

// fetchHTML performs an HTTP GET to rawURL and returns the response body as a
// string. It sets browser-spoofing headers. Go's transport automatically
// negotiates gzip and decompresses the response transparently.
func fetchHTML(rawURL string, opts *Options) (string, error) {
	client := buildHTTPClient(opts)

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("error fetching event, make sure your URL is correct and the event is accessible to the public")
	}

	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching event, make sure your URL is correct and the event is accessible to the public")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching event, make sure your URL is correct and the event is accessible to the public")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error fetching event, make sure your URL is correct and the event is accessible to the public")
	}

	return string(body), nil
}

// buildHTTPClient constructs an http.Client from the given Options.
// Falls back to sane defaults when opts is nil.
func buildHTTPClient(opts *Options) *http.Client {
	transport := http.DefaultTransport
	if opts != nil && opts.Transport != nil {
		transport = opts.Transport
	}

	timeout := 30 * time.Second
	if opts != nil && opts.Timeout > 0 {
		timeout = opts.Timeout
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
