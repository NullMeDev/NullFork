package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"enhanced-gateway-scraper/internal/browser"
	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/gateway"
	"enhanced-gateway-scraper/internal/proxy"
	"enhanced-gateway-scraper/internal/urlgen"
	"enhanced-gateway-scraper/pkg/types"
)

type CLIConfig struct {
	Query       string
	Categories  []string
	ProxyType   string
	ProxyFile   string
	Limit       int
	OutputFile  string
	Format      string
	Headless    bool
	Timeout     int
	UserAgent   string
	Workers     int
}

func main() {
	var cfg CLIConfig

	// Define command line flags
	flag.StringVar(&cfg.Query, "query", "", "Custom search query")
	flag.StringVar(&cfg.Categories, "categories", "payment", "Comma-separated list of categories (payment,proxy,ide,games,sports,ai,crypto,vpn,development,hosting,ecommerce,shopping,fintech,api,cloud)")
	flag.StringVar(&cfg.ProxyType, "proxy-type", "none", "Proxy type: none, http, https, socks4, socks5")
	flag.StringVar(&cfg.ProxyFile, "proxy-file", "", "Path to proxy list file")
	flag.IntVar(&cfg.Limit, "limit", 50, "Maximum number of sites to scan")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output file path (default: stdout)")
	flag.StringVar(&cfg.Format, "format", "json", "Output format: json, table, csv")
	flag.BoolVar(&cfg.Headless, "headless", true, "Run browser in headless mode")
	flag.IntVar(&cfg.Timeout, "timeout", 30, "Browser timeout in seconds")
	flag.StringVar(&cfg.UserAgent, "user-agent", "", "Custom user agent string")
	flag.IntVar(&cfg.Workers, "workers", 10, "Number of concurrent workers")

	flag.Parse()

	// Show help if no arguments
	if len(os.Args) == 1 {
		showUsage()
		return
	}

	// Validate and run
	if err := runEnhancedScraper(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runEnhancedScraper(cfg CLIConfig) error {
	log.Printf("🚀 Enhanced Gateway Scraper starting...")
	log.Printf("📊 Configuration: Categories=%v, Limit=%d, Workers=%d", cfg.Categories, cfg.Limit, cfg.Workers)

	// Load application config
	appConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize components
	urlGenerator := urlgen.NewURLGenerator()
	proxyManager := proxy.NewProxyManager(&types.CheckerConfig{
		MaxWorkers:     cfg.Workers,
		ProxyTimeout:   cfg.Timeout * 1000, // Convert to milliseconds
		RequestTimeout: cfg.Timeout,
	})

	// Set up browser automation
	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = urlGenerator.GetRandomUserAgent()
	}

	browserEngine := browser.NewAutomationEngine(userAgent, cfg.Headless, time.Duration(cfg.Timeout)*time.Second)

	// Load proxies if specified
	if cfg.ProxyFile != "" {
		proxies, err := loadProxiesFromFile(cfg.ProxyFile)
		if err != nil {
			log.Printf("⚠️ Failed to load proxies from file: %v", err)
		} else {
			proxyManager.AddProxies(proxies)
			log.Printf("✅ Loaded %d proxies from file", len(proxies))
		}
	}

	// Scrape additional proxies if needed
	if cfg.ProxyType != "none" && len(proxyManager.GetWorkingProxies()) == 0 {
		log.Printf("🔄 Scraping additional proxies...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		scrapedProxies, err := proxyManager.ScrapeAndValidate(ctx)
		if err != nil {
			log.Printf("⚠️ Failed to scrape proxies: %v", err)
		} else {
			proxyManager.AddProxies(scrapedProxies)
			log.Printf("✅ Scraped and validated %d proxies", len(scrapedProxies))
		}
	}

	// Generate target URLs
	var targetURLs []string
	
	if cfg.Query != "" {
		// Custom query mode
		engines := []string{"Google", "Bing", "DuckDuckGo"}
		targetURLs = urlGenerator.GenerateCustomQuery(cfg.Query, engines)
		log.Printf("🔍 Generated %d URLs from custom query: %s", len(targetURLs), cfg.Query)
	} else {
		// Category-based mode
		categories := parseCategories(cfg.Categories)
		targetURLs = urlGenerator.GenerateSearchURLs(categories, cfg.Limit)
		
		// Add direct URLs for known gateways
		directURLs := urlGenerator.GenerateDirectURLs()
		targetURLs = append(targetURLs, directURLs...)
		
		log.Printf("🎯 Generated %d URLs for categories: %v", len(targetURLs), categories)
	}

	// Limit URLs to specified limit
	if len(targetURLs) > cfg.Limit {
		targetURLs = targetURLs[:cfg.Limit]
	}

	// Initialize gateway detector with rules
	detector := gateway.NewDetector(appConfig.GatewayRules)

	// Scan websites concurrently
	results := scanWebsites(targetURLs, browserEngine, proxyManager, detector, cfg.Workers)

	// Output results
	return outputResults(results, cfg.OutputFile, cfg.Format)
}

func parseCategories(categoriesStr string) []urlgen.Category {
	var categories []urlgen.Category
	parts := strings.Split(categoriesStr, ",")
	
	for _, part := range parts {
		category := strings.TrimSpace(part)
		categories = append(categories, urlgen.Category(category))
	}
	
	return categories
}

func loadProxiesFromFile(filename string) ([]types.Proxy, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var proxies []types.Proxy
	// Simple format: host:port:type (e.g., 127.0.0.1:8080:http)
	// You can extend this to support more complex formats
	
	return proxies, nil
}

func scanWebsites(urls []string, browserEngine *browser.AutomationEngine, proxyManager *proxy.ProxyManager, detector *gateway.Detector, workers int) []ScanResult {
	var results []ScanResult
	var resultsMutex sync.Mutex
	
	// Create work channel
	urlChan := make(chan string, len(urls))
	var wg sync.WaitGroup

	// Fill channel with URLs
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for url := range urlChan {
				result := scanSingleWebsite(url, browserEngine, proxyManager, detector, workerID)
				
				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()
				
				log.Printf("Worker %d: Scanned %s - Found %d gateways", workerID, url, len(result.Gateways))
			}
		}(i)
	}

	wg.Wait()
	return results
}

func scanSingleWebsite(url string, browserEngine *browser.AutomationEngine, proxyManager *proxy.ProxyManager, detector *gateway.Detector, workerID int) ScanResult {
	start := time.Now()
	
	result := ScanResult{
		URL:       url,
		ScanTime:  start,
		Success:   false,
		Gateways:  []types.Gateway{},
		Error:     "",
		Duration:  0,
		WorkerID:  workerID,
	}

	// Get proxy if available
	proxy := proxyManager.GetNextProxy()
	if proxy != nil && proxy.Working {
		proxyURL := fmt.Sprintf("%s://%s:%d", string(proxy.Type), proxy.Host, proxy.Port)
		browserEngine.SetProxy(proxyURL)
		result.ProxyUsed = proxyURL
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Scan website using browser automation
	websiteData, err := browserEngine.ScanWebsite(ctx, url)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// Analyze for gateways using browser automation
	automationGateways := browserEngine.AnalyzeGateways(websiteData)
	
	// Also use traditional detector for comparison
	domain := extractDomainFromURL(url)
	traditionalGateways, err := detector.DetectGateways(ctx, domain)
	if err != nil {
		log.Printf("Traditional detection failed for %s: %v", url, err)
	}

	// Combine results and deduplicate
	allGateways := append(automationGateways, traditionalGateways...)
	result.Gateways = deduplicateGateways(allGateways)
	result.Success = true
	result.Duration = time.Since(start)
	result.PageTitle = websiteData.Title
	result.ScriptCount = len(websiteData.Scripts)
	result.FormCount = len(websiteData.FormData)

	return result
}

func extractDomainFromURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	
	return url
}

func deduplicateGateways(gateways []types.Gateway) []types.Gateway {
	seen := make(map[string]bool)
	var result []types.Gateway
	
	for _, gateway := range gateways {
		key := gateway.Domain + ":" + gateway.GatewayName
		if !seen[key] {
			seen[key] = true
			result = append(result, gateway)
		}
	}
	
	return result
}

func outputResults(results []ScanResult, outputFile, format string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(results, "", "  ")
	case "table":
		output = []byte(formatAsTable(results))
	case "csv":
		output = []byte(formatAsCSV(results))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return err
	}

	if outputFile != "" {
		// Ensure output directory exists
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
		
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Print(string(output))
	return nil
}

func formatAsTable(results []ScanResult) string {
	var sb strings.Builder
	
	sb.WriteString("URL\tStatus\tGateways\tDuration\tProxy\tTitle\n")
	sb.WriteString(strings.Repeat("-", 80) + "\n")
	
	for _, result := range results {
		status := "❌ FAILED"
		if result.Success {
			status = "✅ SUCCESS"
		}
		
		gatewayNames := make([]string, len(result.Gateways))
		for i, gw := range result.Gateways {
			gatewayNames[i] = gw.GatewayName
		}
		
		proxy := "None"
		if result.ProxyUsed != "" {
			proxy = result.ProxyUsed
		}
		
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%v\t%s\t%s\n",
			result.URL,
			status,
			strings.Join(gatewayNames, ","),
			result.Duration.Round(time.Millisecond),
			proxy,
			result.PageTitle,
		))
	}
	
	return sb.String()
}

func formatAsCSV(results []ScanResult) string {
	var sb strings.Builder
	
	sb.WriteString("URL,Status,Gateways,Duration,Proxy,Title,Scripts,Forms,Error\n")
	
	for _, result := range results {
		status := "FAILED"
		if result.Success {
			status = "SUCCESS"
		}
		
		gatewayNames := make([]string, len(result.Gateways))
		for i, gw := range result.Gateways {
			gatewayNames[i] = gw.GatewayName
		}
		
		proxy := ""
		if result.ProxyUsed != "" {
			proxy = result.ProxyUsed
		}
		
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%v,%s,%s,%d,%d,%s\n",
			result.URL,
			status,
			strings.Join(gatewayNames, ";"),
			result.Duration.Milliseconds(),
			proxy,
			result.PageTitle,
			result.ScriptCount,
			result.FormCount,
			result.Error,
		))
	}
	
	return sb.String()
}

func showUsage() {
	fmt.Println("Enhanced Gateway Scraper - Comprehensive Payment Gateway Detection Tool")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  enhanced-scraper [options]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -query string        Custom search query")
	fmt.Println("  -categories string   Categories to scan (default: payment)")
	fmt.Println("                       Available: payment,proxy,ide,games,sports,ai,crypto,vpn,development,hosting,ecommerce,shopping,fintech,api,cloud")
	fmt.Println("  -proxy-type string   Proxy type: none, http, https, socks4, socks5 (default: none)")
	fmt.Println("  -proxy-file string   Path to proxy list file")
	fmt.Println("  -limit int           Maximum sites to scan (default: 50)")
	fmt.Println("  -output string       Output file path (default: stdout)")
	fmt.Println("  -format string       Output format: json, table, csv (default: json)")
	fmt.Println("  -headless            Run browser in headless mode (default: true)")
	fmt.Println("  -timeout int         Browser timeout in seconds (default: 30)")
	fmt.Println("  -user-agent string   Custom user agent")
	fmt.Println("  -workers int         Number of concurrent workers (default: 10)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Scan payment gateways with default settings")
	fmt.Println("  enhanced-scraper -categories payment -limit 20")
	fmt.Println()
	fmt.Println("  # Scan multiple categories with proxy support")
	fmt.Println("  enhanced-scraper -categories payment,crypto,ecommerce -proxy-type http -workers 5")
	fmt.Println()
	fmt.Println("  # Custom search query with table output")
	fmt.Println("  enhanced-scraper -query \"stripe checkout integration\" -format table")
	fmt.Println()
	fmt.Println("  # Save results to file with proxy list")
	fmt.Println("  enhanced-scraper -categories payment -proxy-file proxies.txt -output results.json")
}

// Data structures for CLI results

type ScanResult struct {
	URL         string          `json:"url"`
	ScanTime    time.Time       `json:"scan_time"`
	Success     bool            `json:"success"`
	Gateways    []types.Gateway `json:"gateways"`
	Error       string          `json:"error,omitempty"`
	Duration    time.Duration   `json:"duration"`
	WorkerID    int             `json:"worker_id"`
	ProxyUsed   string          `json:"proxy_used,omitempty"`
	PageTitle   string          `json:"page_title,omitempty"`
	ScriptCount int             `json:"script_count"`
	FormCount   int             `json:"form_count"`
}
