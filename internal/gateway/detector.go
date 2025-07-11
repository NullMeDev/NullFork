package gateway

import (
	"context"
	"log"
	"strings"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/chromedp/chromedp"
	"github.com/valyala/fasthttp"
)

type Detector struct {
	Config *types.CheckerConfig
}

func NewDetector(config *types.CheckerConfig) *Detector {
	return &Detector{
		Config: config,
	}
}

func (d *Detector) DetectGateways(ctx context.Context, domain string) ([]types.Gateway, error) {
	log.Printf("Scanning domain: %s for payment gateways", domain)

	// Prepare context for headless browser
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Target URL
	targetURL := "https://" + domain

	// Run tasks
	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.InnerHTML(`html`, &res),
	)
	if err != nil {
		return nil, err
	}

	// Detect gateways in the HTML content
	gateways := d.extractGateways(res)

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

