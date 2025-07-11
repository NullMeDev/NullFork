package types

import (
	"net/http"
	"time"
)

// ProxyType represents different proxy types
type ProxyType string

const (
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeHTTPS  ProxyType = "https"
	ProxyTypeSOCKS4 ProxyType = "socks4"
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

// Proxy represents a proxy server with health metrics
type Proxy struct {
	ID           string    `json:"id" db:"id"`
	Host         string    `json:"host" db:"host"`
	Port         int       `json:"port" db:"port"`
	Type         ProxyType `json:"type" db:"type"`
	Username     string    `json:"username,omitempty" db:"username"`
	Password     string    `json:"password,omitempty" db:"password"`
	Country      string    `json:"country,omitempty" db:"country"`
	City         string    `json:"city,omitempty" db:"city"`
	ISP          string    `json:"isp,omitempty" db:"isp"`
	Anonymity    string    `json:"anonymity,omitempty" db:"anonymity"`
	Working      bool      `json:"working" db:"working"`
	Latency      int       `json:"latency" db:"latency"`          // milliseconds
	QualityScore int       `json:"quality_score" db:"quality_score"` // 0-100
	LastTest     time.Time `json:"last_test" db:"last_test"`
	FailCount    int       `json:"fail_count" db:"fail_count"`
	SuccessCount int       `json:"success_count" db:"success_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ProxySource represents a source for proxy scraping
type ProxySource struct {
	Name         string            `json:"name" yaml:"name"`
	URL          string            `json:"url" yaml:"url"`
	Type         ProxyType         `json:"type" yaml:"type"`
	Format       string            `json:"format" yaml:"format"` // "ip:port", "ip:port:user:pass", etc.
	Headers      map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Reliability  float64           `json:"reliability" yaml:"reliability"` // 0.0-1.0
	RateLimit    time.Duration     `json:"rate_limit" yaml:"rate_limit"`
	Enabled      bool              `json:"enabled" yaml:"enabled"`
	LastScrape   time.Time         `json:"last_scrape" yaml:"last_scrape"`
	SuccessCount int               `json:"success_count" yaml:"success_count"`
	ErrorCount   int               `json:"error_count" yaml:"error_count"`
}

// Gateway represents a detected payment gateway
type Gateway struct {
	ID             string            `json:"id" db:"id"`
	Domain         string            `json:"domain" db:"domain"`
	URL            string            `json:"url" db:"url"`
	GatewayType    string            `json:"gateway_type" db:"gateway_type"`
	GatewayName    string            `json:"gateway_name" db:"gateway_name"`
	Confidence     float64           `json:"confidence" db:"confidence"` // 0.0-1.0
	DetectionMethod string           `json:"detection_method" db:"detection_method"`
	Patterns       []string          `json:"patterns" db:"patterns"`
	Metadata       map[string]string `json:"metadata" db:"metadata"`
	Screenshot     string            `json:"screenshot,omitempty" db:"screenshot"`
	StatusCode     int               `json:"status_code" db:"status_code"`
	ResponseSize   int               `json:"response_size" db:"response_size"`
	LoadTime       int               `json:"load_time" db:"load_time"` // milliseconds
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
	LastChecked    time.Time         `json:"last_checked" db:"last_checked"`
}

// GatewayRule represents detection rules for payment gateways
type GatewayRule struct {
	Name           string   `json:"name" yaml:"name"`
	Type           string   `json:"type" yaml:"type"`
	Patterns       []string `json:"patterns" yaml:"patterns"`
	DOMSelectors   []string `json:"dom_selectors,omitempty" yaml:"dom_selectors,omitempty"`
	URLPatterns    []string `json:"url_patterns,omitempty" yaml:"url_patterns,omitempty"`
	HeaderPatterns []string `json:"header_patterns,omitempty" yaml:"header_patterns,omitempty"`
	Confidence     float64  `json:"confidence" yaml:"confidence"`
	RequireJS      bool     `json:"require_js" yaml:"require_js"`
	Category       string   `json:"category" yaml:"category"`
	Description    string   `json:"description" yaml:"description"`
}

// Combo represents login credentials for account checking
type Combo struct {
	ID       string `json:"id" db:"id"`
	Line     string `json:"line" db:"line"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Email    string `json:"email,omitempty" db:"email"`
	Domain   string `json:"domain,omitempty" db:"domain"`
}

// Config represents a checking configuration
type Config struct {
	ID             string                 `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Type           string                 `json:"type" db:"type"` // opk, svb, loli
	URL            string                 `json:"url" db:"url"`
	Method         string                 `json:"method" db:"method"`
	Headers        map[string]string      `json:"headers" db:"headers"`
	Data           map[string]interface{} `json:"data" db:"data"`
	SuccessStrings []string               `json:"success_strings" db:"success_strings"`
	FailureStrings []string               `json:"failure_strings" db:"failure_strings"`
	SuccessStatus  []int                  `json:"success_status" db:"success_status"`
	FailureStatus  []int                  `json:"failure_status" db:"failure_status"`
	UseProxy       bool                   `json:"use_proxy" db:"use_proxy"`
	Timeout        time.Duration          `json:"timeout" db:"timeout"`
	MaxRetries     int                    `json:"max_retries" db:"max_retries"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// CheckResult represents the result of a single account check
type CheckResult struct {
	ID        string    `json:"id" db:"id"`
	Combo     Combo     `json:"combo" db:"combo"`
	Config    string    `json:"config" db:"config"`
	Status    string    `json:"status" db:"status"` // valid, invalid, error, unknown
	Response  string    `json:"response,omitempty" db:"response"`
	Error     string    `json:"error,omitempty" db:"error"`
	Proxy     *Proxy    `json:"proxy,omitempty" db:"proxy"`
	Latency   int       `json:"latency" db:"latency"` // milliseconds
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// WorkerTask represents a task for worker processing
type WorkerTask struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // proxy_fetch, proxy_validate, gateway_scan, account_check
	Combo  Combo  `json:"combo,omitempty"`
	Config Config `json:"config,omitempty"`
	Proxy  *Proxy `json:"proxy,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// WorkerResult represents the result of worker processing
type WorkerResult struct {
	TaskID string      `json:"task_id"`
	Type   string      `json:"type"`
	Result CheckResult `json:"result,omitempty"`
	Error  error       `json:"error,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// CheckerConfig represents configuration for the checker engine
type CheckerConfig struct {
	MaxWorkers        int           `json:"max_workers" yaml:"max_workers"`
	ProxyTimeout      int           `json:"proxy_timeout" yaml:"proxy_timeout"`
	RequestTimeout    int           `json:"request_timeout" yaml:"request_timeout"`
	RetryCount        int           `json:"retry_count" yaml:"retry_count"`
	ProxyRotation     bool          `json:"proxy_rotation" yaml:"proxy_rotation"`
	AutoScrapeProxies bool          `json:"auto_scrape_proxies" yaml:"auto_scrape_proxies"`
	SaveValidOnly     bool          `json:"save_valid_only" yaml:"save_valid_only"`
	OutputFormat      string        `json:"output_format" yaml:"output_format"`
	OutputDirectory   string        `json:"output_directory" yaml:"output_directory"`
	RateLimit         time.Duration `json:"rate_limit" yaml:"rate_limit"`
	UserAgent         string        `json:"user_agent" yaml:"user_agent"`
}

// CheckerStats represents runtime statistics
type CheckerStats struct {
	StartTime        time.Time `json:"start_time"`
	ElapsedTime      int       `json:"elapsed_time"`
	TotalCombos      int       `json:"total_combos"`
	ValidCombos      int       `json:"valid_combos"`
	InvalidCombos    int       `json:"invalid_combos"`
	ErrorCombos      int       `json:"error_combos"`
	CurrentCPM       float64   `json:"current_cpm"`
	ActiveWorkers    int       `json:"active_workers"`
	TotalProxies     int       `json:"total_proxies"`
	WorkingProxies   int       `json:"working_proxies"`
	TotalGateways    int       `json:"total_gateways"`
	ProcessedDomains int       `json:"processed_domains"`
}

// ScanSession represents a scanning session
type ScanSession struct {
	ID              string            `json:"id" db:"id"`
	Type            string            `json:"type" db:"type"` // proxy, gateway, checker
	Status          string            `json:"status" db:"status"` // running, completed, error, stopped
	Config          map[string]interface{} `json:"config" db:"config"`
	Stats           CheckerStats      `json:"stats" db:"stats"`
	StartTime       time.Time         `json:"start_time" db:"start_time"`
	EndTime         *time.Time        `json:"end_time,omitempty" db:"end_time"`
	ErrorMessage    string            `json:"error_message,omitempty" db:"error_message"`
	ResultsPath     string            `json:"results_path,omitempty" db:"results_path"`
}

// APIResponse represents standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIMeta represents metadata for API responses
type APIMeta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// WebSocketMessage represents websocket communication
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthStatus represents application health
type HealthStatus struct {
	Status     string            `json:"status"` // healthy, unhealthy, degraded
	Version    string            `json:"version"`
	Uptime     time.Duration     `json:"uptime"`
	Components map[string]string `json:"components"`
	Metrics    map[string]interface{} `json:"metrics"`
	Timestamp  time.Time         `json:"timestamp"`
}

// NotificationConfig represents Discord notification settings
type NotificationConfig struct {
	WebhookURL      string        `json:"webhook_url" yaml:"webhook_url"`
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	ReportInterval  time.Duration `json:"report_interval" yaml:"report_interval"`
	AlertThresholds map[string]float64 `json:"alert_thresholds" yaml:"alert_thresholds"`
	Channels        map[string]string  `json:"channels" yaml:"channels"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type         string            `json:"type" yaml:"type"` // clickhouse, sqlite, postgresql
	DSN          string            `json:"dsn" yaml:"dsn"`
	MaxConns     int               `json:"max_conns" yaml:"max_conns"`
	MaxIdleConns int               `json:"max_idle_conns" yaml:"max_idle_conns"`
	Options      map[string]string `json:"options" yaml:"options"`
}

// HTTPClient interface for testability
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body interface{}) (*http.Response, error)
}
