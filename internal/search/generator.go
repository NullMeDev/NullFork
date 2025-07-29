package search

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// URLGenerator generates URLs based on search queries and categories
type URLGenerator struct {
	searchEngines []SearchEngine
	categories    map[string]CategoryData
	userAgents    []string
}

// SearchEngine represents a search engine configuration
type SearchEngine struct {
	Name    string
	BaseURL string
	Query   string
}

// CategoryData contains keywords and specific sites for each category
type CategoryData struct {
	Keywords []string
	Sites    []string
}

// NewURLGenerator creates a new URL generator with predefined search engines and categories
func NewURLGenerator() *URLGenerator {
	return &URLGenerator{
		searchEngines: []SearchEngine{
			{Name: "Google", BaseURL: "https://www.google.com", Query: "/search?q=%s"},
			{Name: "Bing", BaseURL: "https://www.bing.com", Query: "/search?q=%s"},
			{Name: "DuckDuckGo", BaseURL: "https://duckduckgo.com", Query: "/?q=%s"},
			{Name: "Yahoo", BaseURL: "https://search.yahoo.com", Query: "/search?p=%s"},
		},
		categories: map[string]CategoryData{
			"payment": {
				Keywords: []string{"payment gateway", "payment processor", "checkout", "stripe", "paypal", "square"},
				Sites: []string{
					"https://stripe.com",
					"https://paypal.com",
					"https://square.com",
					"https://checkout.com",
					"https://adyen.com",
					"https://braintreepayments.com",
					"https://razorpay.com",
					"https://klarna.com",
					"https://worldpay.com",
					"https://authorize.net",
				},
			},
			"crypto": {
				Keywords: []string{"cryptocurrency", "bitcoin", "blockchain", "crypto exchange", "wallet"},
				Sites: []string{
					"https://coinbase.com",
					"https://binance.com",
					"https://kraken.com",
					"https://bitpay.com",
					"https://blockchain.com",
					"https://gemini.com",
					"https://crypto.com",
					"https://coindesk.com",
				},
			},
			"ecommerce": {
				Keywords: []string{"ecommerce", "online store", "shopping cart", "marketplace"},
				Sites: []string{
					"https://shopify.com",
					"https://amazon.com",
					"https://ebay.com",
					"https://etsy.com",
					"https://woocommerce.com",
					"https://magento.com",
					"https://bigcommerce.com",
					"https://prestashop.com",
					"https://walmart.com",
					"https://target.com",
					"https://bestbuy.com",
					"https://newegg.com",
					"https://overstock.com",
					"https://wayfair.com",
					"https://alibaba.com",
					"https://aliexpress.com",
					"https://wish.com",
					"https://mercadolibre.com",
					"https://rakuten.com",
				},
			},
			"fintech": {
				Keywords: []string{"fintech", "financial technology", "banking", "investment", "lending"},
				Sites: []string{
					"https://robinhood.com",
					"https://revolut.com",
					"https://wise.com",
					"https://n26.com",
					"https://chime.com",
					"https://plaid.com",
					"https://yodlee.com",
					"https://mint.com",
				},
			},
			"proxy": {
				Keywords: []string{"proxy", "vpn", "proxy service", "socks proxy", "http proxy"},
				Sites: []string{
					"https://nordvpn.com",
					"https://expressvpn.com",
					"https://proxysite.com",
					"https://hideme.ru",
					"https://protonvpn.com",
					"https://surfshark.com",
					"https://cyberghostvpn.com",
					"https://privatevpn.com",
				},
			},
			"vpn": {
				Keywords: []string{"vpn", "virtual private network", "secure connection", "privacy"},
				Sites: []string{
					"https://nordvpn.com",
					"https://expressvpn.com",
					"https://surfshark.com",
					"https://cyberghostvpn.com",
					"https://protonvpn.com",
					"https://privatevpn.com",
					"https://tunnelbear.com",
					"https://windscribe.com",
				},
			},
			"ide": {
				Keywords: []string{"ide", "code editor", "development tools", "programming"},
				Sites: []string{
					"https://code.visualstudio.com",
					"https://jetbrains.com",
					"https://atom.io",
					"https://brackets.io",
					"https://notepad-plus-plus.org",
					"https://vim.org",
					"https://eclipse.org",
					"https://netbeans.apache.org",
				},
			},
			"ai": {
				Keywords: []string{"artificial intelligence", "machine learning", "ai platform", "chatbot"},
				Sites: []string{
					"https://openai.com",
					"https://claude.ai",
					"https://github.com/features/copilot",
					"https://tensorflow.org",
					"https://pytorch.org",
					"https://huggingface.co",
					"https://anthropic.com",
					"https://cohere.ai",
				},
			},
			"games": {
				Keywords: []string{"games", "gaming", "video games", "game platform"},
				Sites: []string{
					"https://steam.com",
					"https://epicgames.com",
					"https://twitch.tv",
					"https://discord.com",
					"https://blizzard.com",
					"https://ubisoft.com",
					"https://ea.com",
					"https://xbox.com",
				},
			},
			"sports": {
				Keywords: []string{"sports", "football", "basketball", "soccer", "baseball"},
				Sites: []string{
					"https://espn.com",
					"https://nfl.com",
					"https://nba.com",
					"https://mlb.com",
					"https://fifa.com",
					"https://olympics.com",
					"https://bleacherreport.com",
					"https://sportsnet.ca",
				},
			},
			"hosting": {
				Keywords: []string{"web hosting", "cloud hosting", "vps", "dedicated server"},
				Sites: []string{
					"https://aws.amazon.com",
					"https://cloud.google.com",
					"https://azure.microsoft.com",
					"https://digitalocean.com",
					"https://linode.com",
					"https://vultr.com",
					"https://godaddy.com",
					"https://bluehost.com",
				},
			},
			"api": {
				Keywords: []string{"api", "rest api", "graphql", "web services", "microservices"},
				Sites: []string{
					"https://rapidapi.com",
					"https://postman.com",
					"https://swagger.io",
					"https://github.com",
					"https://gitlab.com",
					"https://heroku.com",
					"https://netlify.com",
					"https://vercel.com",
				},
			},
		},
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
		},
	}
}

// GenerateFromCategories generates URLs for specified categories
func (ug *URLGenerator) GenerateFromCategories(categories []string, limit int) ([]string, error) {
	var urls []string
	urlsPerCategory := limit / len(categories)
	if urlsPerCategory == 0 {
		urlsPerCategory = 1
	}

	for _, category := range categories {
		categoryData, exists := ug.categories[category]
		if !exists {
			continue
		}

		// Prioritize direct website URLs (80% of URLs should be direct sites)
		directSiteCount := int(float64(urlsPerCategory) * 0.8)
		if directSiteCount == 0 {
			directSiteCount = 1
		}

		// Add specific sites for this category
		for i, site := range categoryData.Sites {
			if len(urls) >= limit || i >= directSiteCount {
				break
			}
			urls = append(urls, site)
		}

		// Add a few search URLs for discovery (20% of URLs)
		searchCount := urlsPerCategory - directSiteCount
		for i, keyword := range categoryData.Keywords {
			if len(urls) >= limit || i >= searchCount {
				break
			}
			
			// Pick a random search engine
			engine := ug.searchEngines[rand.Intn(len(ug.searchEngines))]
			searchURL := fmt.Sprintf(engine.BaseURL+engine.Query, strings.ReplaceAll(keyword, " ", "+"))
			urls = append(urls, searchURL)
		}
	}

	// Shuffle the URLs
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// Ensure we don't exceed the limit
	if len(urls) > limit {
		urls = urls[:limit]
	}

	return urls, nil
}

// GenerateFromQuery generates URLs for a custom search query
func (ug *URLGenerator) GenerateFromQuery(query string, limit int) ([]string, error) {
	var urls []string
	
	// Generate search URLs for each search engine
	for _, engine := range ug.searchEngines {
		if len(urls) >= limit {
			break
		}
		searchURL := fmt.Sprintf(engine.BaseURL+engine.Query, strings.ReplaceAll(query, " ", "+"))
		urls = append(urls, searchURL)
	}

	// Add some variations of the query
	queryVariations := []string{
		query + " website",
		query + " official",
		query + " platform",
		query + " service",
		query + " tool",
	}

	for _, variation := range queryVariations {
		if len(urls) >= limit {
			break
		}
		
		// Pick a random search engine
		engine := ug.searchEngines[rand.Intn(len(ug.searchEngines))]
		searchURL := fmt.Sprintf(engine.BaseURL+engine.Query, strings.ReplaceAll(variation, " ", "+"))
		urls = append(urls, searchURL)
	}

	// Shuffle the URLs
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// Ensure we don't exceed the limit
	if len(urls) > limit {
		urls = urls[:limit]
	}

	return urls, nil
}

// GetRandomUserAgent returns a random user agent string
func (ug *URLGenerator) GetRandomUserAgent() string {
	return ug.userAgents[rand.Intn(len(ug.userAgents))]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
