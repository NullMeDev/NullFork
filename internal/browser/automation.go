package browser

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
)

// AutomationConfig holds configuration for browser automation
type AutomationConfig struct {
	UserAgent string
	Headless  bool
	Timeout   time.Duration
	ProxyURL  string
}

// AutomationEngine handles legitimate browser automation for gateway detection
type AutomationEngine struct {
	config AutomationConfig
}

// ScanResult represents the result of a website scan
type ScanResult struct {
	URL               string                    `json:"url"`
	DetectedGateways  map[string]float64       `json:"detected_gateways"`
	WebsiteData       *WebsiteData             `json:"website_data"`
	ScanDuration      time.Duration            `json:"scan_duration"`
	Error             string                   `json:"error,omitempty"`
}

// NewAutomationEngine creates a new browser automation engine
func NewAutomationEngine(config AutomationConfig) *AutomationEngine {
	return &AutomationEngine{
		config: config,
	}
}

// SetProxy sets the proxy URL for the automation engine
func (ae *AutomationEngine) SetProxy(proxyURL string) {
	ae.config.ProxyURL = proxyURL
}

func (ae *AutomationEngine) ScanWebsite(ctx context.Context, url string) (*ScanResult, error) {
	startTime := time.Now()
	log.Printf("🔍 Starting legitimate browser automation scan for: %s", url)
	
	// Prepare Chrome context with legitimate options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", ae.config.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", false), // Keep web security enabled
		chromedp.UserAgent(ae.config.UserAgent),
	)

	// Add proxy if configured
	if ae.config.ProxyURL != "" {
		opts = append(opts, chromedp.ProxyServer(ae.config.ProxyURL))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, ae.config.Timeout)
	defer cancel()

	// Scan data structure
	websiteData := &WebsiteData{
		URL:              url,
		ScanTime:         time.Now(),
		Scripts:          []ScriptInfo{},
		NetworkRequests:  []NetworkRequest{},
		PaymentElements:  []PaymentElement{},
		FormData:         []FormInfo{},
		MetaTags:         []MetaTag{},
	}

	// Enable network domain
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return nil, fmt.Errorf("failed to enable network: %v", err)
	}

	// Set up network event listener
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			websiteData.NetworkRequests = append(websiteData.NetworkRequests, NetworkRequest{
				URL:    ev.Request.URL,
				Method: ev.Request.Method,
				Headers: convertHeaders(ev.Request.Headers),
			})
		}
	})

	// Navigation and data extraction
	err := chromedp.Run(ctx,
		// Navigate to target
		chromedp.Navigate(url),
		
		// Wait for page load
		chromedp.Sleep(3*time.Second),
		
		// Extract page title
		chromedp.Title(&websiteData.Title),
		
		// Extract meta tags
		ae.extractMetaTags(websiteData),
		
		// Extract all scripts
		ae.extractScripts(websiteData),
		
		// Extract forms
		ae.extractForms(websiteData),
		
		// Extract payment-related elements
		ae.extractPaymentElements(websiteData),
		
		// Extract search result URLs if this is a search engine page
		ae.extractSearchResultURLs(websiteData),
		
		// Extract DOM content
		chromedp.OuterHTML("html", &websiteData.HTMLContent),
	)

	if err != nil {
		return &ScanResult{
			URL:              url,
			DetectedGateways: make(map[string]float64),
			WebsiteData:      websiteData,
			ScanDuration:     time.Since(startTime),
			Error:            err.Error(),
		}, nil
	}

	// Analyze the collected data for payment gateways
	gateways := ae.AnalyzeGateways(websiteData)
	detectedGateways := make(map[string]float64)
	
	for _, gateway := range gateways {
		detectedGateways[gateway.GatewayName] = gateway.Confidence
	}

	// Detailed logging for gateway detection
	if len(detectedGateways) > 0 {
		log.Printf("🎯 PAYMENT GATEWAYS DETECTED on %s:", url)
		for gatewayName, confidence := range detectedGateways {
			log.Printf("  💳 %s (confidence: %.1f%%)", strings.ToUpper(gatewayName), confidence*100)
		}
	} else {
		log.Printf("❌ No payment gateways detected on %s", url)
	}
	
	log.Printf("✅ Scan completed for %s - Found %d scripts, %d forms, %d payment elements, %d gateways", 
		url, len(websiteData.Scripts), len(websiteData.FormData), len(websiteData.PaymentElements), len(detectedGateways))

	return &ScanResult{
		URL:              url,
		DetectedGateways: detectedGateways,
		WebsiteData:      websiteData,
		ScanDuration:     time.Since(startTime),
	}, nil
}

// extractMetaTags extracts meta tags from the page
func (ae *AutomationEngine) extractMetaTags(data *WebsiteData) chromedp.Action {
	return chromedp.Evaluate(`
		Array.from(document.querySelectorAll('meta')).map(meta => ({
			name: meta.getAttribute('name') || meta.getAttribute('property') || '',
			content: meta.getAttribute('content') || '',
			httpEquiv: meta.getAttribute('http-equiv') || ''
		}))
	`, &data.MetaTags)
}

// extractScripts extracts all script information from the page
func (ae *AutomationEngine) extractScripts(data *WebsiteData) chromedp.Action {
	return chromedp.Evaluate(`
		Array.from(document.querySelectorAll('script')).map(script => ({
			src: script.src || '',
			type: script.type || '',
			content: script.innerHTML.substring(0, 500), // Limit content for analysis
			async: script.async || false,
			defer: script.defer || false
		}))
	`, &data.Scripts)
}

// extractForms extracts form information from the page
func (ae *AutomationEngine) extractForms(data *WebsiteData) chromedp.Action {
	return chromedp.Evaluate(`
		Array.from(document.querySelectorAll('form')).map(form => ({
			action: form.action || '',
			method: form.method || 'GET',
			id: form.id || '',
			className: form.className || '',
			fields: Array.from(form.querySelectorAll('input, select, textarea')).map(field => ({
				name: field.name || '',
				type: field.type || '',
				id: field.id || '',
				placeholder: field.placeholder || '',
				required: field.required || false
			}))
		}))
	`, &data.FormData)
}

// extractPaymentElements looks for payment-related elements on the page
func (ae *AutomationEngine) extractPaymentElements(data *WebsiteData) chromedp.Action {
	return chromedp.Evaluate(`
		const paymentSelectors = [
			'[data-stripe]', '[id*="stripe"]', '[class*="stripe"]',
			'[data-paypal]', '[id*="paypal"]', '[class*="paypal"]',
			'[data-square]', '[id*="square"]', '[class*="square"]',
			'[data-braintree]', '[id*="braintree"]', '[class*="braintree"]',
			'[data-adyen]', '[id*="adyen"]', '[class*="adyen"]',
			'[data-razorpay]', '[id*="razorpay"]', '[class*="razorpay"]',
			'[data-checkout]', '[id*="checkout"]', '[class*="checkout"]',
			'[data-payment]', '[id*="payment"]', '[class*="payment"]',
			'[data-gateway]', '[id*="gateway"]', '[class*="gateway"]'
		];
		
		const elements = [];
		paymentSelectors.forEach(selector => {
			try {
				const found = document.querySelectorAll(selector);
				found.forEach(el => {
					elements.push({
						selector: selector,
						tagName: el.tagName,
						id: el.id || '',
						className: el.className || '',
						innerHTML: el.innerHTML.substring(0, 200)
					});
				});
			} catch (e) {
				// Ignore invalid selectors
			}
		});
		
		return elements;
	`, &data.PaymentElements)
}

// extractSearchResultURLs extracts actual website URLs from search engine results
func (ae *AutomationEngine) extractSearchResultURLs(data *WebsiteData) chromedp.Action {
	return chromedp.Evaluate(`
		// Check if this is a search engine page
		const isSearchEngine = window.location.hostname.includes('google.com') ||
							   window.location.hostname.includes('bing.com') ||
							   window.location.hostname.includes('yahoo.com') ||
							   window.location.hostname.includes('duckduckgo.com');
		
		if (!isSearchEngine) {
			return [];
		}
		
		const searchResults = [];
		
		// Google search results
		if (window.location.hostname.includes('google.com')) {
			const results = document.querySelectorAll('a[href*="/url?q="]');
			results.forEach(link => {
				const href = link.getAttribute('href');
				if (href && href.includes('/url?q=')) {
					const url = new URLSearchParams(href.split('?')[1]).get('q');
					if (url && url.startsWith('http')) {
						searchResults.push({
							url: url,
							title: link.textContent || '',
							source: 'google'
						});
					}
				}
			});
		}
		
		// Bing search results
		if (window.location.hostname.includes('bing.com')) {
			const results = document.querySelectorAll('li.b_algo h2 a');
			results.forEach(link => {
				const href = link.getAttribute('href');
				if (href && href.startsWith('http')) {
					searchResults.push({
						url: href,
						title: link.textContent || '',
						source: 'bing'
					});
				}
			});
		}
		
		// Yahoo search results  
		if (window.location.hostname.includes('yahoo.com')) {
			const results = document.querySelectorAll('h3.title a');
			results.forEach(link => {
				const href = link.getAttribute('href');
				if (href && href.startsWith('http')) {
					searchResults.push({
						url: href,
						title: link.textContent || '',
						source: 'yahoo'
					});
				}
			});
		}
		
		// DuckDuckGo search results
		if (window.location.hostname.includes('duckduckgo.com')) {
			const results = document.querySelectorAll('a[data-testid="result-title-a"]');
			results.forEach(link => {
				const href = link.getAttribute('href');
				if (href && href.startsWith('http')) {
					searchResults.push({
						url: href,
						title: link.textContent || '',
						source: 'duckduckgo'
					});
				}
			});
		}
		
		return searchResults.slice(0, 10); // Limit to first 10 results
	`, &data.SearchResults)
}

// AnalyzeGateways analyzes website data to detect payment gateways
func (ae *AutomationEngine) AnalyzeGateways(data *WebsiteData) []types.Gateway {
	var gateways []types.Gateway
	detected := make(map[string]bool)

	// Analyze scripts for gateway patterns
	for _, script := range data.Scripts {
		if gateway := ae.analyzeScript(script, data.URL); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
		}
	}

	// Analyze network requests
	for _, request := range data.NetworkRequests {
		if gateway := ae.analyzeNetworkRequest(request, data.URL); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
		}
	}

	// Analyze payment elements
	for _, element := range data.PaymentElements {
		if gateway := ae.analyzePaymentElement(element, data.URL); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
		}
	}

	// Analyze forms for payment patterns
	for _, form := range data.FormData {
		if gateway := ae.analyzeForm(form, data.URL); gateway != nil && !detected[gateway.GatewayName] {
			gateways = append(gateways, *gateway)
			detected[gateway.GatewayName] = true
		}
	}

	return gateways
}

// analyzeScript analyzes script information for gateway detection
func (ae *AutomationEngine) analyzeScript(script ScriptInfo, baseURL string) *types.Gateway {
	gatewayPatterns := map[string][]string{
		"stripe": {"js.stripe.com", "stripe.network", "checkout.stripe.com", "stripe.com/v3", "stripe.js", "Stripe(", "stripe-checkout"},
		"paypal": {"paypal.com/sdk", "paypalobjects.com", "checkout.paypal.com", "paypal-checkout", "paypal.checkout", "braintree-web"},
		"square": {"js.squareup.com", "web-payments-sdk", "squareup.com", "sq-payment-form", "square-payment"},
		"braintree": {"js.braintreegateway.com", "assets.braintreegateway.com", "braintree-web", "braintree.create"},
		"adyen": {"checkoutshopper-live.adyen.com", "checkoutshopper-test.adyen.com", "adyen.com", "AdyenCheckout"},
		"razorpay": {"checkout.razorpay.com", "razorpay.com", "Razorpay(", "rzp_"},
		"klarna": {"x.klarnacdn.net", "klarna.com", "Klarna.Payments", "klarna-payments"},
		"coinbase": {"commerce.coinbase.com", "coinbase-commerce", "coinbase.com/api"},
		"authorize.net": {"authorize.net", "authorizenet", "acceptjs", "accept.js"},
		"worldpay": {"worldpay.com", "secure.worldpay.com", "worldpay.js"},
		"2checkout": {"2checkout.com", "2co.com", "checkout-2co"},
		"mollie": {"mollie.com", "js.mollie.com", "mollie.js"},
		"recurly": {"js.recurly.com", "recurly.js", "recurly.com"},
		"chargebee": {"js.chargebee.com", "chargebee.js", "chargebee.com"},
		"paddle": {"cdn.paddle.com", "paddle.js", "paddle.com"},
	}

	scriptSource := strings.ToLower(script.Src + " " + script.Content)
	
	for gateway, patterns := range gatewayPatterns {
		for _, pattern := range patterns {
			if strings.Contains(scriptSource, strings.ToLower(pattern)) {
				return &types.Gateway{
					Domain:          extractDomain(baseURL),
					URL:             baseURL,
					GatewayName:     gateway,
					GatewayType:     "payment_processor",
					DetectionMethod: "script_analysis",
					Confidence:      0.9,
					Patterns:        []string{pattern},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					LastChecked:     time.Now(),
				}
			}
		}
	}

	return nil
}

// analyzeNetworkRequest analyzes network requests for gateway APIs
func (ae *AutomationEngine) analyzeNetworkRequest(request NetworkRequest, baseURL string) *types.Gateway {
	apiPatterns := map[string][]string{
		"stripe": {"api.stripe.com", "stripe.com/v1", "stripe.com/v3"},
		"paypal": {"api.paypal.com", "paypal.com/v1", "paypal.com/v2"},
		"square": {"squareup.com/v2", "connect.squareup.com"},
		"adyen": {"checkout-test.adyen.com", "checkout-live.adyen.com"},
		"braintree": {"api.braintreegateway.com", "client-api.braintreegateway.com"},
		"razorpay": {"api.razorpay.com", "razorpay.com/v1"},
	}

	requestURL := strings.ToLower(request.URL)
	
	for gateway, patterns := range apiPatterns {
		for _, pattern := range patterns {
			if strings.Contains(requestURL, pattern) {
				return &types.Gateway{
					Domain:          extractDomain(baseURL),
					URL:             baseURL,
					GatewayName:     gateway,
					GatewayType:     "payment_api",
					DetectionMethod: "network_analysis",
					Confidence:      0.95,
					Patterns:        []string{pattern},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					LastChecked:     time.Now(),
				}
			}
		}
	}

	return nil
}

// analyzePaymentElement analyzes DOM elements for payment gateway indicators
func (ae *AutomationEngine) analyzePaymentElement(element PaymentElement, baseURL string) *types.Gateway {
	elementText := strings.ToLower(element.ID + " " + element.ClassName + " " + element.InnerHTML)
	
	gatewayIndicators := map[string][]string{
		"stripe": {"stripe", "sk_", "pk_"},
		"paypal": {"paypal", "paypal-checkout"},
		"square": {"square", "square-payment"},
		"braintree": {"braintree", "bt-"},
		"adyen": {"adyen", "adyen-checkout"},
		"razorpay": {"razorpay", "rzp_"},
	}

	for gateway, indicators := range gatewayIndicators {
		for _, indicator := range indicators {
			if strings.Contains(elementText, indicator) {
				return &types.Gateway{
					Domain:          extractDomain(baseURL),
					URL:             baseURL,
					GatewayName:     gateway,
					GatewayType:     "payment_element",
					DetectionMethod: "dom_analysis",
					Confidence:      0.8,
					Patterns:        []string{indicator},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					LastChecked:     time.Now(),
				}
			}
		}
	}

	return nil
}

// analyzeForm analyzes form data for payment processing indicators
func (ae *AutomationEngine) analyzeForm(form FormInfo, baseURL string) *types.Gateway {
	formData := strings.ToLower(form.Action + " " + form.ID + " " + form.ClassName)
	
	// Check form fields for payment indicators
	for _, field := range form.Fields {
		fieldData := strings.ToLower(field.Name + " " + field.ID + " " + field.Placeholder)
		formData += " " + fieldData
	}

	paymentIndicators := []string{"payment", "checkout", "card", "billing", "stripe", "paypal"}
	
	for _, indicator := range paymentIndicators {
		if strings.Contains(formData, indicator) {
			return &types.Gateway{
				Domain:          extractDomain(baseURL),
				URL:             baseURL,
				GatewayName:     "generic_payment_form",
				GatewayType:     "payment_form",
				DetectionMethod: "form_analysis",
				Confidence:      0.7,
				Patterns:        []string{indicator},
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				LastChecked:     time.Now(),
			}
		}
	}

	return nil
}

// Helper functions

func convertHeaders(headers map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

func extractDomain(urlStr string) string {
	if strings.HasPrefix(urlStr, "http://") {
		urlStr = urlStr[7:]
	} else if strings.HasPrefix(urlStr, "https://") {
		urlStr = urlStr[8:]
	}
	
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	
	return urlStr
}

// Data structures for website analysis

type WebsiteData struct {
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	HTMLContent      string            `json:"html_content"`
	Scripts          []ScriptInfo      `json:"scripts"`
	NetworkRequests  []NetworkRequest  `json:"network_requests"`
	PaymentElements  []PaymentElement  `json:"payment_elements"`
	FormData         []FormInfo        `json:"forms"`
	MetaTags         []MetaTag         `json:"meta_tags"`
	SearchResults    []SearchResult    `json:"search_results"`
	ScanTime         time.Time         `json:"scan_time"`
}

type ScriptInfo struct {
	Src     string `json:"src"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Async   bool   `json:"async"`
	Defer   bool   `json:"defer"`
}

type NetworkRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type PaymentElement struct {
	Selector  string `json:"selector"`
	TagName   string `json:"tag_name"`
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	InnerHTML string `json:"inner_html"`
}

type FormInfo struct {
	Action    string      `json:"action"`
	Method    string      `json:"method"`
	ID        string      `json:"id"`
	ClassName string      `json:"class_name"`
	Fields    []FieldInfo `json:"fields"`
}

type FieldInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ID          string `json:"id"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
}

type MetaTag struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	HttpEquiv string `json:"http_equiv"`
}

type SearchResult struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Source string `json:"source"`
}
