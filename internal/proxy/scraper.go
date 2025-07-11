package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/valyala/fasthttp"
)

// Scraper manages proxy scraping and validation
type Scraper struct {
	Config *types.CheckerConfig
}

// NewScraper creates a new Scraper
func NewScraper(config *types.CheckerConfig) *Scraper {
	return &Scraper{
		Config: config,
	}
}

// ScrapeAndValidate scrapes and validates proxies
func (s *Scraper) ScrapeAndValidate(ctx context.Context) ([]types.Proxy, error) {
	var wg sync.WaitGroup
	proxyChan := make(chan types.Proxy, 1000)

	// Define proxy scraping sources
	sources := []string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://www.proxy-list.download/api/v1/get?type=https",
		"https://www.proxy-list.download/api/v1/get?type=socks4",
		"https://www.proxy-list.download/api/v1/get?type=socks5",
	}

	// Launch goroutines to fetch proxies
	for _, source := range sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			s.fetchFromSource(src, proxyChan)
		}(source)
	}

	// Close channel when done
	go func() {
		wg.Wait()
		close(proxyChan)
	}()

	// Collect and validate proxies
	var proxies []types.Proxy
	for proxy := range proxyChan {
		if s.validateProxy(ctx, proxy) {
			proxies = append(proxies, proxy)
		}
	}

	return proxies, nil
}

// fetchFromSource scrapes proxies from a given source
func (s *Scraper) fetchFromSource(source string, proxyChan chan<- types.Proxy) {
	client := fasthttp.Client{
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	statusCode, body, err := client.Get(nil, source)
	if err != nil || statusCode != fasthttp.StatusOK {
		return
	}

	// Parse proxies from response body
	// Example assumes proxies are in ip:port format per line
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), ":")
		if len(fields) == 2 {
			proxy := types.Proxy{
				Host: fields[0],
				Port: s.parsePort(fields[1]),
				Type: s.detectProxyType(source),
			}
			proxyChan <- proxy
		}
	}
}

// validateProxy verifies if a proxy is valid
func (s *Scraper) validateProxy(ctx context.Context, proxy types.Proxy) bool {
	// Create HTTP client with proxy
	proxyURL := fmt.Sprintf("%s://%s:%d", string(proxy.Type), proxy.Host, proxy.Port)
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: time.Duration(s.Config.ProxyTimeout) * time.Millisecond,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Test proxy connectivity
	start := time.Now()
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Update proxy metrics
	proxy.Latency = int(time.Since(start).Milliseconds())
	proxy.Working = resp.StatusCode == 200
	proxy.LastTest = time.Now()

	if proxy.Working {
		proxy.SuccessCount++
		// Calculate quality score based on latency and success rate
		proxy.QualityScore = s.calculateQualityScore(proxy)
	} else {
		proxy.FailCount++
	}

	return proxy.Working && proxy.QualityScore >= s.Config.ProxyTimeout
}

// detectProxyType detects the proxy type from the source URL
func (s *Scraper) detectProxyType(source string) types.ProxyType {
	switch {
	case strings.Contains(source, "socks4"):
		return types.ProxyTypeSOCKS4
	case strings.Contains(source, "socks5"):
		return types.ProxyTypeSOCKS5
	case strings.Contains(source, "https"):
		return types.ProxyTypeHTTPS
	default:
		return types.ProxyTypeHTTP
	}
}

// parsePort safely parses port from string
func (s *Scraper) parsePort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

// calculateQualityScore calculates proxy quality based on performance metrics
func (s *Scraper) calculateQualityScore(proxy types.Proxy) int {
	// Base score
	score := 100

	// Penalize high latency
	if proxy.Latency > 5000 {
		score -= 50
	} else if proxy.Latency > 2000 {
		score -= 30
	} else if proxy.Latency > 1000 {
		score -= 15
	}

	// Factor in success rate
	totalTests := proxy.SuccessCount + proxy.FailCount
	if totalTests > 0 {
		successRate := float64(proxy.SuccessCount) / float64(totalTests)
		score = int(float64(score) * successRate)
	}

	// Ensure score is within valid range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
