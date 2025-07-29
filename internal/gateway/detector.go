package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
)

type Detector struct {
	Rules map[string]types.GatewayRule
}

func NewDetector(rules map[string]types.GatewayRule) *Detector {
	return &Detector{
		Rules: rules,
	}
}

func (d *Detector) DetectGateways(ctx context.Context, domain string) ([]types.Gateway, error) {
	log.Printf("🔍 Advanced scanning domain: %s for payment gateways", domain)

	// Prepare context for headless browser with advanced options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Target URL
	targetURL := "https://" + domain

	// Advanced detection variables
	var (
		htmlContent    string
		scriptSources  []string
		networkRequests []string
		localStorage   map[string]string
		cookieData     []*http.Cookie
	)

	// Run comprehensive scanning tasks
	err := chromedp.Run(ctx,
		// Navigate and wait for page load
		chromedp.Navigate(targetURL),
		chromedp.Sleep(3*time.Second), // Wait for dynamic content
		
		// Extract HTML content
		chromedp.InnerHTML(`html`, &htmlContent),
		
		// Extract all script sources
		chromedp.Evaluate(`Array.from(document.querySelectorAll('script[src]')).map(s => s.src)`, &scriptSources),
		
		// Extract network requests from performance API
		chromedp.Evaluate(`
			Array.from(performance.getEntriesByType('resource'))
				.map(entry => entry.name)
				.filter(url => url.includes('js') || url.includes('payment') || url.includes('checkout'))
		`, &networkRequests),
		
		// Check localStorage for payment-related data
		chromedp.Evaluate(`
			let storage = {};
			for(let i = 0; i < localStorage.length; i++) {
				let key = localStorage.key(i);
				if(key.toLowerCase().includes('payment') || 
				   key.toLowerCase().includes('stripe') ||
				   key.toLowerCase().includes('paypal') ||
				   key.toLowerCase().includes('gateway')) {
					storage[key] = localStorage.getItem(key);
				}
			}
			storage;
		`, &localStorage),
	)

	if err != nil {
		// Fallback to basic HTTP request if browser fails
		log.Printf("⚠️ Browser scan failed, falling back to HTTP: %v", err)
		return d.fallbackDetection(domain)
	}

	// Get cookies
	chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		networks, _ := network.GetCookies().Do(ctx)
		for _, cookie := range networks {
			cookieData = append(cookieData, &http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			})
		}
		return nil
	}))

	// Advanced gateway detection
	gateways := d.advancedGatewayExtraction(htmlContent, scriptSources, networkRequests, localStorage, cookieData)

	log.Printf("✅ Scan complete: found %d gateways", len(gateways))
	return gateways, nil
}

func (d *Detector) extractGateways(htmlContent string) []types.Gateway {
	var gateways []types.Gateway

	// Basic pattern matching for known gateway scripts/URLs
	patterns := map[string]string{
		"stripe":          "js.stripe.com",
		"paypal":          "www.paypal.com/sdk/js",
		"square":          "js.squareup.com",
		"authorize-net":   "secure.authorize.net/gateway/transact.dll",
		"braintree":       "assets.braintreegateway.com/web/dropin/",
		"razorpay":        "checkout.razorpay.com/v1/checkout.js",
		"mollie":          "js.mollie.com",
		"adyen":           "checkoutshopper-live.adyen.com",
	}

	for name, pattern := range patterns {
		if strings.Contains(htmlContent, pattern) {
			gateways = append(gateways, types.Gateway{
				GatewayName: name,
				DetectionMethod: "script_pattern",
				Confidence: 0.9,
				Patterns: []string{pattern},
			})
		}
	}

	return gateways
}

// Advanced gateway extraction using multiple detection methods
func (d *Detector) advancedGatewayExtraction(htmlContent string, scriptSources []string, networkRequests []string, localStorage map[string]string, cookies []*http.Cookie) []types.Gateway {
	var gateways []types.Gateway
	detected := make(map[string]bool) // Prevent duplicates

	log.Printf("🧠 Running advanced detection on %d scripts, %d network requests", len(scriptSources), len(networkRequests))

	// Method 1: Enhanced script source analysis
	for _, script := range scriptSources {
		if gateway := d.analyzeScriptSource(script); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
			log.Printf("🎯 Found %s via script analysis: %s", gateway.GatewayName, script)
		}
	}

	// Method 2: Network request analysis
	for _, request := range networkRequests {
		if gateway := d.analyzeNetworkRequest(request); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
			log.Printf("🌐 Found %s via network analysis: %s", gateway.GatewayName, request)
		}
	}

	// Method 3: HTML content pattern matching
	for ruleName, rule := range d.Rules {
		if !detected[ruleName] {
			if gateway := d.analyzeHTMLContent(htmlContent, ruleName, rule); gateway != nil {
				gateways = append(gateways, *gateway)
				detected[gateway.GatewayName] = true
				log.Printf("📄 Found %s via HTML analysis", gateway.GatewayName)
			}
		}
	}

	// Method 4: Cookie analysis
	for _, cookie := range cookies {
		if gateway := d.analyzeCookie(cookie); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
			log.Printf("🍪 Found %s via cookie analysis: %s", gateway.GatewayName, cookie.Name)
		}
	}

	// Method 5: localStorage analysis
	for key, value := range localStorage {
		if gateway := d.analyzeLocalStorage(key, value); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
			log.Printf("💾 Found %s via localStorage analysis: %s", gateway.GatewayName, key)
		}
	}

	return gateways
}

// Deep script source analysis
func (d *Detector) analyzeScriptSource(scriptURL string) *types.Gateway {
	// Advanced script patterns with higher precision
	advancedPatterns := map[string][]string{
		"stripe": {
			"js.stripe.com",
			"checkout.stripe.com",
			"m.stripe.com",
			"stripe.network",
		},
		"paypal": {
			"paypal.com/sdk",
			"paypalobjects.com",
			"paypal-sdk",
			"checkout.paypal.com",
		},
		"square": {
			"js.squareup.com",
			"web-payments-sdk",
			"squareup.com",
			"square-web-payments-sdk",
		},
		"braintree": {
			"js.braintreegateway.com",
			"assets.braintreegateway.com",
			"braintree-web",
		},
		"adyen": {
			"checkoutshopper-live.adyen.com",
			"checkoutshopper-test.adyen.com",
			"adyen.com/library",
		},
		"authorize-net": {
			"secure.authorize.net",
			"jstest.authorize.net",
			"js.authorize.net",
		},
		"razorpay": {
			"checkout.razorpay.com",
			"razorpay.com/checkout",
		},
		"klarna": {
			"x.klarnacdn.net",
			"klarna.com",
			"klarna-payments",
		},
		"coinbase": {
			"commerce.coinbase.com",
			"coinbase-commerce",
		},
		"shopify": {
			"checkout.shopifycs.com",
			"shop.app",
			"shopify-pay",
		},
	}

	for gateway, patterns := range advancedPatterns {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(scriptURL), strings.ToLower(pattern)) {
				return &types.Gateway{
					GatewayName:     gateway,
					DetectionMethod: "advanced_script_analysis",
					Confidence:      0.95,
					Patterns:        []string{pattern},
				}
			}
		}
	}

	return nil
}

// Network request analysis
func (d *Detector) analyzeNetworkRequest(requestURL string) *types.Gateway {
	// Look for payment-related API calls
	paymentAPIs := map[string][]string{
		"stripe": {"api.stripe.com", "stripe.com/v1", "stripe.com/v3"},
		"paypal": {"api.paypal.com", "paypal.com/v1", "paypal.com/v2"},
		"square": {"squareup.com/v2", "connect.squareup.com"},
		"adyen": {"checkout-test.adyen.com", "checkout-live.adyen.com"},
		"braintree": {"api.braintreegateway.com", "client-api.braintreegateway.com"},
	}

	for gateway, apis := range paymentAPIs {
		for _, api := range apis {
			if strings.Contains(strings.ToLower(requestURL), api) {
				return &types.Gateway{
					GatewayName:     gateway,
					DetectionMethod: "network_request_analysis",
					Confidence:      0.90,
					Patterns:        []string{api},
				}
			}
		}
	}

	return nil
}

// HTML content deep analysis using regex patterns
func (d *Detector) analyzeHTMLContent(htmlContent, ruleName string, rule types.GatewayRule) *types.Gateway {
	// Enhanced pattern matching using regex
	for _, pattern := range rule.Patterns {
		// Try exact match first
		if strings.Contains(strings.ToLower(htmlContent), strings.ToLower(pattern)) {
			return &types.Gateway{
				GatewayName:     ruleName,
				DetectionMethod: "html_pattern_match",
				Confidence:      rule.Confidence,
				Patterns:        []string{pattern},
			}
		}

		// Try regex pattern matching
		if regex, err := regexp.Compile(pattern); err == nil {
			if regex.MatchString(htmlContent) {
				return &types.Gateway{
					GatewayName:     ruleName,
					DetectionMethod: "html_regex_match",
					Confidence:      rule.Confidence * 0.95, // Slightly lower confidence for regex
					Patterns:        []string{pattern},
				}
			}
		}
	}

	return nil
}

// Cookie analysis for payment gateway detection
func (d *Detector) analyzeCookie(cookie *http.Cookie) *types.Gateway {
	paymentCookiePatterns := map[string][]string{
		"stripe":     {"stripe", "sk_", "pk_"},
		"paypal":     {"paypal", "PYPF", "tsrce"},
		"square":     {"square", "sq-"},
		"braintree":  {"braintree", "bt_"},
		"authorize":  {"authorize", "authnet"},
		"shopify":    {"shopify", "_shopify"},
	}

	for gateway, patterns := range paymentCookiePatterns {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(cookie.Name), pattern) ||
			   strings.Contains(strings.ToLower(cookie.Value), pattern) {
				return &types.Gateway{
					GatewayName:     gateway,
					DetectionMethod: "cookie_analysis",
					Confidence:      0.85,
					Patterns:        []string{pattern},
				}
			}
		}
	}

	return nil
}

// localStorage analysis
func (d *Detector) analyzeLocalStorage(key, value string) *types.Gateway {
	localStoragePatterns := map[string][]string{
		"stripe":    {"stripe", "pk_live", "pk_test"},
		"paypal":    {"paypal", "PAYPAL"},
		"square":    {"square", "APPLICATION_ID"},
		"braintree": {"braintree", "client_token"},
	}

	lowerKey := strings.ToLower(key)
	lowerValue := strings.ToLower(value)

	for gateway, patterns := range localStoragePatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerKey, strings.ToLower(pattern)) ||
			   strings.Contains(lowerValue, strings.ToLower(pattern)) {
				return &types.Gateway{
					GatewayName:     gateway,
					DetectionMethod: "localstorage_analysis",
					Confidence:      0.80,
					Patterns:        []string{pattern},
				}
			}
		}
	}

	return nil
}

// Fallback detection using basic HTTP request
func (d *Detector) fallbackDetection(domain string) ([]types.Gateway, error) {
	log.Printf("🔄 Using fallback HTTP detection for %s", domain)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s", domain))
	if err != nil {
		return nil, fmt.Errorf("fallback request failed: %v", err)
	}
	defer resp.Body.Close()

	// Basic header analysis
	var gateways []types.Gateway

	// Check for payment-related headers
	for name, values := range resp.Header {
		for _, value := range values {
			if gateway := d.analyzeHeader(name, value); gateway != nil {
				gateways = append(gateways, *gateway)
			}
		}
	}

	return gateways, nil
}

// Header analysis for payment gateway detection
func (d *Detector) analyzeHeader(name, value string) *types.Gateway {
	headerPatterns := map[string][]string{
		"stripe":     {"stripe", "Stripe-"},
		"paypal":     {"paypal", "PayPal-"},
		"square":     {"square", "Square-"},
		"shopify":    {"shopify", "X-Shopify"},
	}

	lowerName := strings.ToLower(name)
	lowerValue := strings.ToLower(value)

	for gateway, patterns := range headerPatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerName, strings.ToLower(pattern)) ||
			   strings.Contains(lowerValue, strings.ToLower(pattern)) {
				return &types.Gateway{
					GatewayName:     gateway,
					DetectionMethod: "http_header_analysis",
					Confidence:      0.75,
					Patterns:        []string{pattern},
				}
			}
		}
	}

	return nil
}
