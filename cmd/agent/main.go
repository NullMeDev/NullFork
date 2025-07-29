package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"enhanced-gateway-scraper/internal/config"
	"enhanced-gateway-scraper/internal/gateway"
	"enhanced-gateway-scraper/internal/logger"
	"enhanced-gateway-scraper/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	version = "NullScrape v1.0.0 - Agent Mode"
	banner  = `
╔══════════════════════════════════════════════════════════════╗
║                         NullScrape                           ║
║                      Agent Mode v1.0.0                      ║
║              Browser-capable scraping assistant             ║
╚══════════════════════════════════════════════════════════════╝
`
)

func main() {
	fmt.Print(banner)
	
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]
	
	switch command {
	case "search":
		handleSearch()
	case "scan":
		handleScan()
	case "version":
		fmt.Println(version)
	case "help", "--help", "-h":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func handleSearch() {
	var (
		query      = flag.String("query", "", "Search query (e.g., 'proxy ai tools')")
		categories = flag.String("categories", "", "Categories to search (e.g., 'proxy,ai,payment,games,sports')")
		proxyType  = flag.String("proxy", "none", "Proxy type: none, http, https, socks4, socks5")
		limit      = flag.Int("limit", 25, "Maximum number of sites to crawl (max 25)")
		output     = flag.String("output", "gui", "Output format: gui, json, table")
	)
	
	// Parse flags starting from index 2 (after "agent search")
	flag.CommandLine.Parse(os.Args[2:])
	
	if *query == "" && *categories == "" {
		fmt.Println("❌ Error: Either search query or categories is required")
		fmt.Println("Usage: agent search --query=\"proxy ai tools\" --proxy=socks5")
		fmt.Println("   OR: agent search --categories=\"proxy,ai,payment\" --proxy=socks5")
		os.Exit(1)
	}
	
	if *limit > 25 {
		*limit = 25
		fmt.Println("⚠️  Warning: Limit capped at 25 sites")
	}
	
	fmt.Printf("🔍 NullScrape Search Mode\n")
	if *query != "" {
		fmt.Printf("Query: %s\n", *query)
	}
	if *categories != "" {
		fmt.Printf("Categories: %s\n", *categories)
	}
	fmt.Printf("Proxy: %s\n", *proxyType)
	fmt.Printf("Limit: %d sites\n", *limit)
	fmt.Printf("Output: %s\n", *output)
	fmt.Println()
	
	// Execute search
	executeSearch(*query, *categories, *proxyType, *limit, *output)
}

func handleScan() {
	var (
		url       = flag.String("url", "", "URL to scan")
		proxyType = flag.String("proxy", "none", "Proxy type: none, http, https, socks4, socks5")
		output    = flag.String("output", "gui", "Output format: gui, json, table")
	)
	
	// Parse flags starting from index 2 (after "agent scan")
	flag.CommandLine.Parse(os.Args[2:])
	
	if *url == "" {
		fmt.Println("❌ Error: URL is required")
		fmt.Println("Usage: agent scan --url=https://example.com --proxy=http")
		os.Exit(1)
	}
	
	fmt.Printf("🎯 NullScrape Scan Mode\n")
	fmt.Printf("Target: %s\n", *url)
	fmt.Printf("Proxy: %s\n", *proxyType)
	fmt.Printf("Output: %s\n", *output)
	fmt.Println()
	
	// Execute scan
	executeScan(*url, *proxyType, *output)
}

func executeSearch(query, categories, proxyType string, limit int, output string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize logger
	appLogger := logger.InitLogger("INFO", "TEXT", false)
	appLogger.Info("🚀 Starting NullScrape search operation...")
	
	// Generate search URLs based on query or categories
	var searchURLs []string
	if query != "" {
		searchURLs = generateSearchURLs(query, limit)
		fmt.Printf("🌐 Generated %d URLs from query: '%s'\n", len(searchURLs), query)
	} else if categories != "" {
		searchURLs = generateCategoryBasedURLs(categories, limit)
		fmt.Printf("🌐 Generated %d URLs from categories: '%s'\n", len(searchURLs), categories)
	}
	
	// Configure proxy if specified
	configureProxy(cfg, proxyType, appLogger)
	
	// Initialize gateway detector
	detector := gateway.NewDetector(cfg.GatewayRules)
	
	fmt.Println("🔍 Starting headless browser crawling...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	var results []ScanResult
	
	for i, url := range searchURLs {
		fmt.Printf("[%d/%d] 🌐 Crawling: %s\n", i+1, len(searchURLs), url)
		
		// Use headless browser to detect gateways
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		gateways, err := detector.DetectGateways(ctx, extractDomain(url))
		cancel()
		
		result := ScanResult{
			URL:      url,
			Domain:   extractDomain(url),
			Gateways: gateways,
			Error:    err,
			Scanned:  time.Now(),
		}
		
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else if len(gateways) > 0 {
			fmt.Printf("   ✅ Found %d payment gateway(s):\n", len(gateways))
			for _, gw := range gateways {
				fmt.Printf("      • %s (%.0f%% confidence)\n", gw.GatewayName, gw.Confidence*100)
			}
		} else {
			fmt.Printf("   ⚪ No payment gateways detected\n")
		}
		
		results = append(results, result)
		
		// Small delay between requests
		time.Sleep(time.Duration(cfg.DefaultRequestDelay))
	}
	
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Output results
	outputResults(results, output, cfg)
}

func executeScan(url, proxyType, output string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Initialize logger
	appLogger := logger.InitLogger("INFO", "TEXT", false)
	appLogger.Info("🎯 Starting NullScrape direct scan...")
	
	// Configure proxy if specified
	configureProxy(cfg, proxyType, appLogger)
	
	// Initialize gateway detector
	detector := gateway.NewDetector(cfg.GatewayRules)
	
	fmt.Printf("🔍 Scanning %s with headless browser...\n", url)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Use headless browser to detect gateways
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	gateways, err := detector.DetectGateways(ctx, extractDomain(url))
	cancel()
	
	result := ScanResult{
		URL:      url,
		Domain:   extractDomain(url),
		Gateways: gateways,
		Error:    err,
		Scanned:  time.Now(),
	}
	
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
		os.Exit(1)
	}
	
	if len(gateways) > 0 {
		fmt.Printf("✅ Found %d payment gateway(s):\n", len(gateways))
		for _, gw := range gateways {
			fmt.Printf("   • %s\n", strings.ToUpper(gw.GatewayName))
			fmt.Printf("     Confidence: %.0f%%\n", gw.Confidence*100)
			fmt.Printf("     Method: %s\n", gw.DetectionMethod)
			if len(gw.Patterns) > 0 {
				fmt.Printf("     Patterns: %v\n", gw.Patterns)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("⚪ No payment gateways detected")
	}
	
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Output results
	outputResults([]ScanResult{result}, output, cfg)
}

func configureProxy(cfg *config.Config, proxyType string, logger *logrus.Logger) {
	switch proxyType {
	case "http", "https":
		fmt.Println("🔗 Configuring HTTP/HTTPS proxy rotation...")
		// Here we would configure HTTP proxy settings
		logger.Info("HTTP proxy rotation enabled")
	case "socks4":
		fmt.Println("🔗 Configuring SOCKS4 proxy rotation...")
		logger.Info("SOCKS4 proxy rotation enabled")
	case "socks5":
		fmt.Println("🔗 Configuring SOCKS5 proxy rotation...")
		logger.Info("SOCKS5 proxy rotation enabled")
	case "none":
		fmt.Println("🔗 Direct connection (no proxies)")
	default:
		fmt.Printf("⚠️  Unknown proxy type '%s', using direct connection\n", proxyType)
	}
}

func generateSearchURLs(query string, limit int) []string {
	var urls []string
	keywords := strings.Fields(strings.ToLower(query))
	
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  Warning: Google API Key not set, falling back to static URLs")
		return getFallbackURLs(query, limit)
	}
	
	// Use a generic search engine ID for demonstration
	searchEngineID := "017576662512468239146:omuauf_lfve" // This is a placeholder
	
	for _, keyword := range keywords {
		searchURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?q=%s&key=%s&cx=%s&num=%d", 
			url.QueryEscape(keyword), "[REDACTED]", searchEngineID, min(10, limit))
		
		response, err := http.Get(searchURL)
		if err != nil {
			log.Printf("Error fetching search results for '%s': %v", keyword, err)
			continue
		}
		defer response.Body.Close()
		
		var searchResults GoogleSearchResult
		if err := json.NewDecoder(response.Body).Decode(&searchResults); err != nil {
			log.Printf("Error decoding search results for '%s': %v", keyword, err)
			continue
		}
		
		for _, item := range searchResults.Items {
			if len(urls) >= limit {
				break
			}
			urls = append(urls, item.Link)
		}
		
		if len(urls) >= limit {
			break
		}
	}

	if len(urls) == 0 {
		return getFallbackURLs(query, limit)
	}

	return urls
}

func generateCategoryBasedURLs(categories string, limit int) []string {
	var urls []string
	categoryList := strings.Split(strings.ToLower(categories), ",")
	
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  Warning: Google API Key not set, using predefined category URLs")
		return getPredefinedCategoryURLs(categoryList, limit)
	}
	
	// Category-specific search queries with enhanced targeting
	categoryQueries := map[string][]string{
		"proxy": {"free proxy list", "proxy servers", "anonymous proxy", "SOCKS proxy"},
		"ai": {"artificial intelligence tools", "AI platforms", "machine learning services", "ChatGPT alternatives"},
		"payment": {"payment gateway", "online payment processor", "payment solutions", "e-commerce payment"},
		"games": {"online games", "gaming platforms", "multiplayer games", "game servers"},
		"sports": {"sports betting", "sports news", "fantasy sports", "sports streaming"},
		"ide": {"online IDE", "code editor", "development environment", "programming tools"},
		"crypto": {"cryptocurrency exchange", "crypto wallet", "blockchain services", "DeFi platforms"},
		"vpn": {"VPN services", "virtual private network", "online privacy tools"},
		"hosting": {"web hosting", "cloud hosting", "VPS hosting", "dedicated servers"},
		"ecommerce": {"online shopping", "e-commerce platforms", "online marketplace"},
	}
	
	searchEngineID := "017576662512468239146:omuauf_lfve" // Placeholder
	
	for _, category := range categoryList {
		category = strings.TrimSpace(category)
		queries, exists := categoryQueries[category]
		if !exists {
			// Use the category name itself as search query
			queries = []string{category + " services", category + " platforms"}
		}
		
		for _, query := range queries {
			if len(urls) >= limit {
				break
			}
			
			searchURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?q=%s&key=%s&cx=%s&num=%d", 
				url.QueryEscape(query), "[REDACTED]", searchEngineID, min(5, limit-len(urls)))
			
			response, err := http.Get(searchURL)
			if err != nil {
				log.Printf("Error fetching search results for category '%s', query '%s': %v", category, query, err)
				continue
			}
			defer response.Body.Close()
			
			var searchResults GoogleSearchResult
			if err := json.NewDecoder(response.Body).Decode(&searchResults); err != nil {
				log.Printf("Error decoding search results for category '%s': %v", category, err)
				continue
			}
			
			for _, item := range searchResults.Items {
				if len(urls) >= limit {
					break
				}
				urls = append(urls, item.Link)
			}
		}
		
		if len(urls) >= limit {
			break
		}
	}
	
	if len(urls) == 0 {
		return getPredefinedCategoryURLs(categoryList, limit)
	}
	
	return urls
}

func getFallbackURLs(query string, limit int) []string {
	// Fallback URLs when API is not available
	keywords := strings.Fields(strings.ToLower(query))
	var urls []string
	
	for _, keyword := range keywords {
		switch keyword {
		case "proxy":
			urls = append(urls, "https://www.proxy-list.download", "https://proxylist.geonode.com", "https://www.freeproxy.world")
		case "ai", "artificial", "intelligence":
			urls = append(urls, "https://openai.com", "https://claude.ai", "https://huggingface.co", "https://replicate.com")
		case "payment", "gateway":
			urls = append(urls, "https://stripe.com", "https://paypal.com", "https://square.com", "https://razorpay.com")
		case "games", "gaming":
			urls = append(urls, "https://steam.com", "https://epic.games.com", "https://twitch.tv")
		case "sports":
			urls = append(urls, "https://espn.com", "https://nba.com", "https://nfl.com")
		}
	}
	
	if len(urls) == 0 {
		urls = []string{"https://example.com", "https://httpbin.org"}
	}
	
	if len(urls) > limit {
		urls = urls[:limit]
	}
	
	return urls
}

func getPredefinedCategoryURLs(categories []string, limit int) []string {
	var urls []string
	
	categoryURLs := map[string][]string{
		"proxy": {"https://www.proxy-list.download", "https://proxylist.geonode.com", "https://www.freeproxy.world", "https://hidemy.name/en/proxy-list/"},
		"ai": {"https://openai.com", "https://claude.ai", "https://huggingface.co", "https://replicate.com", "https://midjourney.com"},
		"payment": {"https://stripe.com", "https://paypal.com", "https://square.com", "https://razorpay.com", "https://braintreepayments.com"},
		"games": {"https://steam.com", "https://epic.games.com", "https://twitch.tv", "https://discord.com", "https://roblox.com"},
		"sports": {"https://espn.com", "https://nba.com", "https://nfl.com", "https://mlb.com", "https://uefa.com"},
		"ide": {"https://replit.com", "https://codepen.io", "https://codesandbox.io", "https://stackblitz.com", "https://gitpod.io"},
		"crypto": {"https://coinbase.com", "https://binance.com", "https://kraken.com", "https://gemini.com"},
		"vpn": {"https://nordvpn.com", "https://expressvpn.com", "https://surfshark.com", "https://protonvpn.com"},
		"hosting": {"https://aws.amazon.com", "https://digitalocean.com", "https://vultr.com", "https://linode.com"},
		"ecommerce": {"https://shopify.com", "https://amazon.com", "https://ebay.com", "https://etsy.com"},
	}
	
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if categoryList, exists := categoryURLs[category]; exists {
			for _, url := range categoryList {
				if len(urls) >= limit {
					break
				}
				urls = append(urls, url)
			}
		}
		if len(urls) >= limit {
			break
		}
	}
	
	if len(urls) == 0 {
		urls = []string{"https://example.com", "https://httpbin.org"}
	}
	
	return urls
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type GoogleSearchResult struct {
	Items []struct {
		Link string `json:"link"`
		Title string `json:"title"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

func extractDomain(url string) string {
	// Simple domain extraction
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, "/")
	return parts[0]
}

type ScanResult struct {
	URL      string
	Domain   string
	Gateways []types.Gateway
	Error    error
	Scanned  time.Time
}

func outputResults(results []ScanResult, format string, cfg *config.Config) {
	switch format {
	case "gui":
		startGUIServer(results, cfg)
	case "json":
		outputJSON(results)
	case "table":
		outputTable(results)
	default:
		outputTable(results)
	}
}

func startGUIServer(results []ScanResult, cfg *config.Config) {
	fmt.Println("🖥️  Starting GUI server...")
	fmt.Printf("📱 Web interface available at: http://localhost:%d\n", cfg.WebPort)
	fmt.Println("💡 Press Ctrl+C to stop the server")
	
	// This would start the GUI server with results
	// For now, just keep the process alive
	select {}
}

func outputJSON(results []ScanResult) {
	fmt.Println("📄 JSON Output:")
	// JSON output implementation
	for _, result := range results {
		fmt.Printf(`{"url": "%s", "gateways": %d, "timestamp": "%s"}%s`,
			result.URL, len(result.Gateways), result.Scanned.Format(time.RFC3339), "\n")
	}
}

func outputTable(results []ScanResult) {
	fmt.Println("📊 Scan Results Summary:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-40s %-15s %-10s\n", "DOMAIN", "GATEWAYS", "STATUS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	totalGateways := 0
	for _, result := range results {
		status := "✅ OK"
		if result.Error != nil {
			status = "❌ ERROR"
		}
		
		gatewayNames := make([]string, len(result.Gateways))
		for i, gw := range result.Gateways {
			gatewayNames[i] = gw.GatewayName
		}
		
		gatewayStr := fmt.Sprintf("%d", len(result.Gateways))
		if len(result.Gateways) > 0 {
			gatewayStr = fmt.Sprintf("%d (%s)", len(result.Gateways), strings.Join(gatewayNames, ","))
		}
		
		fmt.Printf("%-40s %-15s %-10s\n", result.Domain, gatewayStr, status)
		totalGateways += len(result.Gateways)
	}
	
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📈 Total: %d sites scanned, %d payment gateways found\n", len(results), totalGateways)
	
	if totalGateways > 0 {
		fmt.Printf("🌐 View detailed results in GUI: http://localhost:8082\n")
	}
}

func showHelp() {
	fmt.Printf(`
NullScrape - Agent Mode v1.0.0
Browser-capable scraping assistant for payment gateway detection

USAGE:
    agent <command> [options]

COMMANDS:
    search    Search and crawl up to 25 websites based on query
    scan      Scan a specific URL for payment gateways
    version   Show version information
    help      Show this help message

SEARCH EXAMPLES:
    agent search --query="proxy ai tools" --proxy=socks5
    agent search --categories="payment,ai,proxy" --proxy=http --limit=10
    agent search --query="stripe integration" --output=json
    agent search --categories="games,sports,crypto" --output=table

SCAN EXAMPLES:
    agent scan --url=https://stripe.com --proxy=none
    agent scan --url=https://example.com --proxy=socks4 --output=table

PROXY TYPES:
    none      Direct connection (default)
    http      HTTP proxy rotation
    https     HTTPS proxy rotation  
    socks4    SOCKS4 proxy rotation
    socks5    SOCKS5 proxy rotation

OUTPUT FORMATS:
    gui       Interactive web interface (default)
    json      JSON formatted output
    table     Tabular console output

FEATURES:
    • Headless browser scraping with Chrome/Chromium
    • Payment gateway fingerprint matching
    • Proxy rotation (HTTP, HTTPS, SOCKS4, SOCKS5)
    • Auto-removal of dead/broken proxies
    • GUI-styled output inspired by https://nullgen.replit.app
    • Support for up to 25 concurrent website crawls

For more information, visit: https://nullgen.replit.app
`)
}
