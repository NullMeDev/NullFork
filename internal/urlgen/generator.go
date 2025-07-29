package urlgen

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"math/rand"
)

// Category represents different target categories for scanning
type Category string

const (
	CategoryPayment     Category = "payment"
	CategoryProxy       Category = "proxy"
	CategoryIDE         Category = "ide"
	CategoryGames       Category = "games"
	CategorySports      Category = "sports"
	CategoryAI          Category = "ai"
	CategoryCrypto      Category = "crypto"
	CategoryVPN         Category = "vpn"
	CategoryDevelopment Category = "development"
	CategoryHosting     Category = "hosting"
	CategoryEcommerce   Category = "ecommerce"
	CategoryShopping    Category = "shopping"
	CategoryFintech     Category = "fintech"
	CategoryAPI         Category = "api"
	CategoryCloud       Category = "cloud"
)

// SearchEngine represents different search engines and their endpoints
type SearchEngine struct {
	Name     string
	BaseURL  string
	QueryKey string
	Headers  map[string]string
}

// URLGenerator handles generation of target URLs for scanning
type URLGenerator struct {
	SearchEngines []SearchEngine
	Categories    map[Category][]string
	UserAgents    []string
}

// NewURLGenerator creates a new URL generator with predefined search engines and categories
func NewURLGenerator() *URLGenerator {
	return &URLGenerator{
		SearchEngines: []SearchEngine{
			{
				Name:     "Google",
				BaseURL:  "https://www.google.com/search",
				QueryKey: "q",
				Headers: map[string]string{
					"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				},
			},
			{
				Name:     "Bing",
				BaseURL:  "https://www.bing.com/search",
				QueryKey: "q",
				Headers: map[string]string{
					"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				},
			},
			{
				Name:     "DuckDuckGo",
				BaseURL:  "https://duckduckgo.com/",
				QueryKey: "q",
				Headers: map[string]string{
					"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
				},
			},
		},
		Categories: map[Category][]string{
			CategoryPayment: {
				"stripe checkout", "paypal payment", "square payment gateway",
				"authorize.net", "braintree payment", "adyen checkout",
				"razorpay", "mollie payment", "worldpay", "payoneer",
				"klarna checkout", "afterpay", "sezzle payment",
			},
			CategoryProxy: {
				"free proxy list", "socks5 proxy", "http proxy",
				"residential proxy", "datacenter proxy", "rotating proxy",
				"proxy api", "proxy service", "anonymous proxy",
			},
			CategoryIDE: {
				"visual studio code", "jetbrains", "atom editor",
				"sublime text", "eclipse ide", "netbeans",
				"code editor", "online ide", "cloud ide",
			},
			CategoryGames: {
				"steam games", "epic games", "battle.net",
				"origin games", "uplay", "gog games",
				"mobile games", "indie games", "gaming platform",
			},
			CategorySports: {
				"sports betting", "fantasy sports", "sports streaming",
				"sports news", "live sports", "sports analytics",
				"sports equipment", "fitness tracking",
			},
			CategoryAI: {
				"ai chatbot", "machine learning", "artificial intelligence",
				"neural network", "deep learning", "ai api",
				"computer vision", "natural language processing",
				"ai tools", "ai platform",
			},
			CategoryCrypto: {
				"bitcoin exchange", "cryptocurrency", "crypto wallet",
				"blockchain", "ethereum", "crypto trading",
				"defi platform", "nft marketplace", "crypto api",
			},
			CategoryVPN: {
				"vpn service", "virtual private network", "secure vpn",
				"vpn provider", "vpn client", "vpn server",
				"wireguard", "openvpn", "vpn protocol",
			},
			CategoryDevelopment: {
				"github", "gitlab", "development tools",
				"api documentation", "sdk", "framework",
				"library", "package manager", "version control",
			},
			CategoryHosting: {
				"web hosting", "cloud hosting", "vps hosting",
				"dedicated server", "domain registration", "cdn",
				"aws", "azure", "google cloud", "digitalocean",
			},
			CategoryEcommerce: {
				"shopify", "woocommerce", "magento",
				"bigcommerce", "prestashop", "opencart",
				"online store", "ecommerce platform",
			},
			CategoryShopping: {
				"amazon", "ebay", "walmart", "target",
				"online shopping", "marketplace", "retail",
				"shopping cart", "product catalog",
			},
			CategoryFintech: {
				"fintech api", "banking api", "payment processor",
				"financial services", "robo advisor", "lending platform",
				"investment platform", "trading platform",
			},
			CategoryAPI: {
				"rest api", "graphql api", "api gateway",
				"api management", "webhook", "microservices",
				"api documentation", "api testing",
			},
			CategoryCloud: {
				"cloud platform", "saas", "paas", "iaas",
				"cloud storage", "cloud computing", "serverless",
				"container", "kubernetes", "docker",
			},
		},
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:122.0) Gecko/20100101 Firefox/122.0",
		},
	}
}

// GenerateSearchURLs creates URLs for searching specific categories
func (g *URLGenerator) GenerateSearchURLs(categories []Category, limit int) []string {
	var urls []string
	rand.Seed(time.Now().UnixNano())

	for _, category := range categories {
		keywords, exists := g.Categories[category]
		if !exists {
			continue
		}

		// Shuffle keywords for randomization
		shuffledKeywords := make([]string, len(keywords))
		copy(shuffledKeywords, keywords)
		rand.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})

		// Generate URLs for each search engine
		for _, engine := range g.SearchEngines {
			for i, keyword := range shuffledKeywords {
				if len(urls) >= limit {
					return urls
				}

				// Create variations of search queries
				queries := g.generateQueryVariations(keyword, category)
				for _, query := range queries {
					if len(urls) >= limit {
						return urls
					}

					searchURL := g.buildSearchURL(engine, query)
					urls = append(urls, searchURL)
				}

				// Limit keywords per engine to avoid too many URLs
				if i >= 3 {
					break
				}
			}
		}
	}

	return urls
}

// GenerateDirectURLs creates direct URLs for known payment gateway domains
func (g *URLGenerator) GenerateDirectURLs() []string {
	directDomains := []string{
		// Payment Gateways
		"stripe.com", "js.stripe.com", "checkout.stripe.com",
		"paypal.com", "www.paypal.com", "checkout.paypal.com",
		"square.com", "squareup.com", "js.squareup.com",
		"authorize.net", "secure.authorize.net",
		"braintreegateway.com", "js.braintreegateway.com",
		"adyen.com", "checkoutshopper-live.adyen.com",
		"razorpay.com", "checkout.razorpay.com",
		"mollie.com", "js.mollie.com",
		"worldpay.com", "secure.worldpay.com",
		"klarna.com", "x.klarnacdn.net",
		
		// Crypto Exchanges
		"coinbase.com", "binance.com", "kraken.com",
		"bitfinex.com", "gemini.com", "crypto.com",
		
		// E-commerce Platforms
		"shopify.com", "checkout.shopifycs.com",
		"woocommerce.com", "magento.com",
		"bigcommerce.com", "prestashop.com",
		
		// Cloud Providers
		"aws.amazon.com", "azure.microsoft.com",
		"cloud.google.com", "digitalocean.com",
		"linode.com", "vultr.com",
		
		// Development Tools
		"github.com", "gitlab.com", "bitbucket.org",
		"npmjs.com", "pypi.org", "packagist.org",
	}

	var urls []string
	for _, domain := range directDomains {
		urls = append(urls, "https://"+domain)
		urls = append(urls, "https://www."+domain)
	}

	return urls
}

// GenerateCustomQuery creates URLs for custom search queries
func (g *URLGenerator) GenerateCustomQuery(query string, engines []string) []string {
	var urls []string

	for _, engineName := range engines {
		for _, engine := range g.SearchEngines {
			if strings.EqualFold(engine.Name, engineName) {
				searchURL := g.buildSearchURL(engine, query)
				urls = append(urls, searchURL)
				break
			}
		}
	}

	return urls
}

// generateQueryVariations creates different variations of search queries
func (g *URLGenerator) generateQueryVariations(keyword string, category Category) []string {
	variations := []string{keyword}

	// Add category-specific modifiers
	switch category {
	case CategoryPayment:
		variations = append(variations,
			keyword+" integration",
			keyword+" api",
			keyword+" checkout",
			keyword+" payment gateway",
			"\""+keyword+"\"",
		)
	case CategoryProxy:
		variations = append(variations,
			keyword+" server",
			keyword+" list",
			keyword+" api",
			"\""+keyword+"\"",
		)
	case CategoryIDE:
		variations = append(variations,
			keyword+" download",
			keyword+" features",
			keyword+" extensions",
			"\""+keyword+"\"",
		)
	case CategoryGames:
		variations = append(variations,
			keyword+" platform",
			keyword+" store",
			keyword+" launcher",
			"\""+keyword+"\"",
		)
	case CategoryAI:
		variations = append(variations,
			keyword+" platform",
			keyword+" service",
			keyword+" api",
			keyword+" tools",
			"\""+keyword+"\"",
		)
	default:
		variations = append(variations,
			keyword+" service",
			keyword+" platform",
			"\""+keyword+"\"",
		)
	}

	// Add site-specific searches for better targeting
	siteSpecific := []string{
		"site:github.com " + keyword,
		"site:stackoverflow.com " + keyword,
		"site:reddit.com " + keyword,
	}
	variations = append(variations, siteSpecific...)

	return variations
}

// buildSearchURL constructs a complete search URL
func (g *URLGenerator) buildSearchURL(engine SearchEngine, query string) string {
	params := url.Values{}
	params.Add(engine.QueryKey, query)
	
	// Add additional parameters for better results
	switch engine.Name {
	case "Google":
		params.Add("num", "50")
		params.Add("hl", "en")
	case "Bing":
		params.Add("count", "50")
		params.Add("mkt", "en-US")
	case "DuckDuckGo":
		params.Add("t", "h_")
		params.Add("ia", "web")
	}

	return engine.BaseURL + "?" + params.Encode()
}

// GetRandomUserAgent returns a random user agent string
func (g *URLGenerator) GetRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return g.UserAgents[rand.Intn(len(g.UserAgents))]
}

// GetCategoryKeywords returns keywords for a specific category
func (g *URLGenerator) GetCategoryKeywords(category Category) []string {
	if keywords, exists := g.Categories[category]; exists {
		return keywords
	}
	return []string{}
}

// GetAllCategories returns all available categories
func (g *URLGenerator) GetAllCategories() []Category {
	var categories []Category
	for category := range g.Categories {
		categories = append(categories, category)
	}
	return categories
}

// ValidateCategory checks if a category is valid
func (g *URLGenerator) ValidateCategory(category string) bool {
	_, exists := g.Categories[Category(category)]
	return exists
}
