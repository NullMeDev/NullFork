package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Proxy represents a proxy server
type Proxy struct {
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Type       string    `json:"type"`
	Working    bool      `json:"working"`
	Latency    int64     `json:"latency"`
	LastUsed   time.Time `json:"last_used"`
	UseCount   int       `json:"use_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProxyManager manages proxy scraping, validation, and rotation
type ProxyManager struct {
	proxies      []*Proxy
	mutex        sync.RWMutex
	currentIndex int
}

// NewProxyManager creates a new ProxyManager
func NewProxyManager() (*ProxyManager, error) {
	return &ProxyManager{
		proxies: make([]*Proxy, 0),
	}, nil
}

// AddProxy adds a single proxy to the manager
func (pm *ProxyManager) AddProxy(proxy *Proxy) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	proxy.CreatedAt = time.Now()
	proxy.UpdatedAt = time.Now()
	pm.proxies = append(pm.proxies, proxy)
}

// GetWorkingProxies returns only working proxies
func (pm *ProxyManager) GetWorkingProxies() []*Proxy {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	var working []*Proxy
	for _, proxy := range pm.proxies {
		if proxy.Working {
			working = append(working, proxy)
		}
	}
	return working
}

// GetRandomProxy returns a random working proxy
func (pm *ProxyManager) GetRandomProxy() *Proxy {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	working := make([]*Proxy, 0)
	for _, proxy := range pm.proxies {
		if proxy.Working {
			working = append(working, proxy)
		}
	}
	
	if len(working) == 0 {
		return nil
	}
	
	// Simple round-robin selection
	pm.currentIndex = (pm.currentIndex + 1) % len(working)
	return working[pm.currentIndex]
}

// ValidateProxies validates all proxies in the manager
func (pm *ProxyManager) ValidateProxies(ctx context.Context) {
	pm.mutex.RLock()
	proxies := make([]*Proxy, len(pm.proxies))
	copy(proxies, pm.proxies)
	pm.mutex.RUnlock()
	
	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		go func(p *Proxy) {
			defer wg.Done()
			pm.validateProxy(ctx, p)
		}(proxy)
	}
	wg.Wait()
}

// ScrapeProxies scrapes proxies from various sources
func (pm *ProxyManager) ScrapeProxies(ctx context.Context) error {
	sources := []string{
		"https://www.proxy-list.download/api/v1/get?type=http",
		"https://www.proxy-list.download/api/v1/get?type=https",
	}
	
	var wg sync.WaitGroup
	for _, source := range sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			pm.scrapeFromSource(ctx, src)
		}(source)
	}
	wg.Wait()
	
	return nil
}

// scrapeFromSource scrapes proxies from a single source
func (pm *ProxyManager) scrapeFromSource(ctx context.Context, source string) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	resp, err := client.Get(source)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return
	}
	
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Count(line, ":") == 1 {
			fields := strings.Split(line, ":")
			if len(fields) != 2 {
				continue
			}
			
			host := strings.TrimSpace(fields[0])
			portStr := strings.TrimSpace(fields[1])
			port, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			
			proxy := &Proxy{
				Host: host,
				Port: port,
				Type: pm.detectProxyType(source),
			}
			
			pm.AddProxy(proxy)
		}
	}
}

// validateProxy validates a single proxy
func (pm *ProxyManager) validateProxy(ctx context.Context, proxy *Proxy) {
	proxyURL := fmt.Sprintf("http://%s:%d", proxy.Host, proxy.Port)
	pURL, err := url.Parse(proxyURL)
	if err != nil {
		proxy.Working = false
		return
	}
	
	transport := &http.Transport{
		Proxy:           http.ProxyURL(pURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	
	start := time.Now()
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		proxy.Working = false
		proxy.Latency = 0
		return
	}
	defer resp.Body.Close()
	
	proxy.Latency = time.Since(start).Milliseconds()
	proxy.Working = resp.StatusCode == 200
	proxy.UpdatedAt = time.Now()
}

// detectProxyType determines proxy type from source URL
func (pm *ProxyManager) detectProxyType(source string) string {
	switch {
	case strings.Contains(source, "socks4"):
		return "socks4"
	case strings.Contains(source, "socks5"):
		return "socks5"
	case strings.Contains(source, "https"):
		return "https"
	default:
		return "http"
	}
}
