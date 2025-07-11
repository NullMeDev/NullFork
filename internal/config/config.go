package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	// Database
	Database types.DatabaseConfig `yaml:"database"`

	// Discord
	Discord types.NotificationConfig `yaml:"discord"`

	// Performance
	MaxWorkers           int `yaml:"max_workers"`
	MaxConcurrentDomains int `yaml:"max_concurrent_domains"`
	MinDomainScanCount   int `yaml:"min_domain_scan_count"`
	MaxDomainScanCount   int `yaml:"max_domain_scan_count"`
	DefaultScanCount     int `yaml:"default_scan_count"`

	// Proxy settings
	MinProxyCount             int           `yaml:"min_proxy_count"`
	MaxProxyCount             int           `yaml:"max_proxy_count"`
	ProxyQualityThreshold     int           `yaml:"proxy_quality_threshold"`
	ProxyValidationWorkers    int           `yaml:"proxy_validation_workers"`
	ProxyTimeout              time.Duration `yaml:"proxy_timeout"`
	ProxyHealthCheckInterval  time.Duration `yaml:"proxy_health_check_interval"`

	// Rate limiting
	EnableRateLimiting      bool                    `yaml:"enable_rate_limiting"`
	DefaultRequestDelay     time.Duration           `yaml:"default_request_delay"`
	GatewayRateLimits       map[string]time.Duration `yaml:"gateway_rate_limits"`

	// Timeouts
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	GatewayScanTimeout    time.Duration `yaml:"gateway_scan_timeout"`
	ProxyValidationTimeout time.Duration `yaml:"proxy_validation_timeout"`

	// User agents
	EnableUserAgentRotation bool     `yaml:"enable_user_agent_rotation"`
	UserAgents              []string `yaml:"user_agents"`

	// Target categories
	TargetCategories []string `yaml:"target_categories"`
	PriorityGateways []string `yaml:"priority_gateways"`

	// Scanning intervals
	ProxyFetchInterval    time.Duration `yaml:"proxy_fetch_interval"`
	ProxyValidateInterval time.Duration `yaml:"proxy_validate_interval"`
	GatewayScanInterval   time.Duration `yaml:"gateway_scan_interval"`
	HealthCheckInterval   time.Duration `yaml:"health_check_interval"`
	DiscordReportInterval time.Duration `yaml:"discord_report_interval"`

	// JavaScript rendering
	EnableJavaScriptRendering bool          `yaml:"enable_javascript_rendering"`
	HeadlessBrowser          bool          `yaml:"headless_browser"`
	BrowserTimeout           time.Duration `yaml:"browser_timeout"`
	MaxRenderTime            time.Duration `yaml:"max_render_time"`

	// Logging
	LogLevel        string `yaml:"log_level"`
	LogFormat       string `yaml:"log_format"`
	LogIncludeCaller bool   `yaml:"log_include_caller"`
	EnableMetrics   bool   `yaml:"enable_metrics"`
	MetricsPort     int    `yaml:"metrics_port"`

	// API & Web
	APIPort       int  `yaml:"api_port"`
	WebPort       int  `yaml:"web_port"`
	EnableAPIAuth bool `yaml:"enable_api_auth"`
	EnableCORS    bool `yaml:"enable_cors"`

	// Storage
	DataDirectory    string        `yaml:"data_directory"`
	BackupDirectory  string        `yaml:"backup_directory"`
	ResultsDirectory string        `yaml:"results_directory"`
	AutoBackupInterval time.Duration `yaml:"auto_backup_interval"`
	MaxBackupRetention time.Duration `yaml:"max_backup_retention"`

	// Security
	EnableProxyGeolocation  bool          `yaml:"enable_proxy_geolocation"`
	RequireProxyAnonymity   bool          `yaml:"require_proxy_anonymity"`
	EnableTLSVerification   bool          `yaml:"enable_tls_verification"`
	EnableRequestRandomization bool       `yaml:"enable_request_randomization"`
	EnableCookiePersistence bool          `yaml:"enable_cookie_persistence"`
	SessionTimeout          time.Duration `yaml:"session_timeout"`

	// Gateway Rules
	GatewayRules map[string]types.GatewayRule `yaml:"gateway_rules"`
	RateLimits   map[string]time.Duration     `yaml:"rate_limits"`
}

// Load loads configuration from environment variables and files
func Load() (*Config, error) {
	config := &Config{}

	// Load from environment variables
	config.loadFromEnv()

	// Load gateway rules from YAML file
	if err := config.loadGatewayRules(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// Database
	c.Database = types.DatabaseConfig{
		Type: "clickhouse",
		DSN:  getEnv("CLICKHOUSE_DSN", "localhost:9000"),
		MaxConns: getEnvInt("CLICKHOUSE_MAX_CONNS", 10),
		MaxIdleConns: getEnvInt("CLICKHOUSE_MAX_IDLE_CONNS", 5),
	}

	// Discord
	c.Discord = types.NotificationConfig{
		WebhookURL:     getEnv("DISCORD_WEBHOOK", ""),
		Enabled:        getEnv("DISCORD_WEBHOOK", "") != "",
		ReportInterval: getEnvDuration("DISCORD_REPORT_INTERVAL", time.Hour),
		AlertThresholds: map[string]float64{
			"error_rate":     getEnvFloat("DISCORD_ERROR_THRESHOLD", 0.1),
			"success_rate":   getEnvFloat("DISCORD_SUCCESS_THRESHOLD", 0.8),
			"proxy_health":   getEnvFloat("DISCORD_PROXY_HEALTH_THRESHOLD", 0.7),
		},
	}

	// Performance
	c.MaxWorkers = getEnvInt("MAX_WORKERS", 100)
	c.MaxConcurrentDomains = getEnvInt("MAX_CONCURRENT_DOMAINS", 500)
	c.MinDomainScanCount = getEnvInt("MIN_DOMAIN_SCAN_COUNT", 5)
	c.MaxDomainScanCount = getEnvInt("MAX_DOMAIN_SCAN_COUNT", 500)
	c.DefaultScanCount = getEnvInt("DEFAULT_DOMAIN_SCAN_COUNT", 50)

	// Proxy settings
	c.MinProxyCount = getEnvInt("MIN_PROXY_COUNT", 5)
	c.MaxProxyCount = getEnvInt("MAX_PROXY_COUNT", 5000)
	c.ProxyQualityThreshold = getEnvInt("PROXY_QUALITY_THRESHOLD", 75)
	c.ProxyValidationWorkers = getEnvInt("PROXY_VALIDATION_WORKERS", 50)
	c.ProxyTimeout = getEnvDuration("PROXY_TIMEOUT", 10*time.Second)
	c.ProxyHealthCheckInterval = getEnvDuration("PROXY_HEALTH_CHECK_INTERVAL", 5*time.Minute)

	// Rate limiting
	c.EnableRateLimiting = getEnvBool("ENABLE_RATE_LIMITING", true)
	c.DefaultRequestDelay = getEnvDuration("DEFAULT_REQUEST_DELAY", time.Second)

	// Timeouts
	c.RequestTimeout = getEnvDuration("REQUEST_TIMEOUT", 30*time.Second)
	c.GatewayScanTimeout = getEnvDuration("GATEWAY_SCAN_TIMEOUT", 45*time.Second)
	c.ProxyValidationTimeout = getEnvDuration("PROXY_VALIDATION_TIMEOUT", 10*time.Second)

	// User agents
	c.EnableUserAgentRotation = getEnvBool("ENABLE_USER_AGENT_ROTATION", true)
	userAgentsStr := getEnv("USER_AGENTS", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	c.UserAgents = strings.Split(userAgentsStr, ",")

	// Target categories
	categoriesStr := getEnv("TARGET_CATEGORIES", "proxies,tools,ide,ai-agents")
	c.TargetCategories = strings.Split(categoriesStr, ",")
	
	gatewaysStr := getEnv("PRIORITY_GATEWAYS", "stripe,paypal,square")
	c.PriorityGateways = strings.Split(gatewaysStr, ",")

	// Scanning intervals
	c.ProxyFetchInterval = getEnvDuration("PROXY_FETCH_INTERVAL", 15*time.Minute)
	c.ProxyValidateInterval = getEnvDuration("PROXY_VALIDATE_INTERVAL", 30*time.Minute)
	c.GatewayScanInterval = getEnvDuration("GATEWAY_SCAN_INTERVAL", 2*time.Hour)
	c.HealthCheckInterval = getEnvDuration("HEALTH_CHECK_INTERVAL", 5*time.Minute)
	c.DiscordReportInterval = getEnvDuration("DISCORD_REPORT_INTERVAL", time.Hour)

	// JavaScript rendering
	c.EnableJavaScriptRendering = getEnvBool("ENABLE_JAVASCRIPT_RENDERING", true)
	c.HeadlessBrowser = getEnvBool("HEADLESS_BROWSER", true)
	c.BrowserTimeout = getEnvDuration("BROWSER_TIMEOUT", 30*time.Second)
	c.MaxRenderTime = getEnvDuration("MAX_RENDER_TIME", 15*time.Second)

	// Logging
	c.LogLevel = getEnv("LOG_LEVEL", "INFO")
	c.LogFormat = getEnv("LOG_FORMAT", "JSON")
	c.LogIncludeCaller = getEnvBool("LOG_INCLUDE_CALLER", true)
	c.EnableMetrics = getEnvBool("ENABLE_METRICS", true)
	c.MetricsPort = getEnvInt("METRICS_PORT", 9090)

	// API & Web
	c.APIPort = getEnvInt("API_PORT", 8080)
	c.WebPort = getEnvInt("WEB_PORT", 8081)
	c.EnableAPIAuth = getEnvBool("ENABLE_API_AUTH", false)
	c.EnableCORS = getEnvBool("ENABLE_CORS", true)

	// Storage
	c.DataDirectory = getEnv("DATA_DIRECTORY", "./data")
	c.BackupDirectory = getEnv("BACKUP_DIRECTORY", "./backups")
	c.ResultsDirectory = getEnv("RESULTS_DIRECTORY", "./results")
	c.AutoBackupInterval = getEnvDuration("AUTO_BACKUP_INTERVAL", 24*time.Hour)
	c.MaxBackupRetention = getEnvDuration("MAX_BACKUP_RETENTION", 30*24*time.Hour)

	// Security
	c.EnableProxyGeolocation = getEnvBool("ENABLE_PROXY_GEOLOCATION", false)
	c.RequireProxyAnonymity = getEnvBool("REQUIRE_PROXY_ANONYMITY", true)
	c.EnableTLSVerification = getEnvBool("ENABLE_TLS_VERIFICATION", false)
	c.EnableRequestRandomization = getEnvBool("ENABLE_REQUEST_RANDOMIZATION", true)
	c.EnableCookiePersistence = getEnvBool("ENABLE_COOKIE_PERSISTENCE", true)
	c.SessionTimeout = getEnvDuration("SESSION_TIMEOUT", time.Hour)
}

// loadGatewayRules loads gateway detection rules from YAML file
func (c *Config) loadGatewayRules() error {
	viper.SetConfigName("gateway-rules")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	c.GatewayRules = make(map[string]types.GatewayRule)
	
	// Load gateway rules
	gatewayRulesMap := viper.GetStringMap("gateway_rules")
	for name, ruleData := range gatewayRulesMap {
		var rule types.GatewayRule
		ruleBytes, _ := yaml.Marshal(ruleData)
		if err := yaml.Unmarshal(ruleBytes, &rule); err != nil {
			continue
		}
		c.GatewayRules[name] = rule
	}

	// Load rate limits
	c.RateLimits = make(map[string]time.Duration)
	rateLimitsMap := viper.GetStringMapString("rate_limits")
	for gateway, limitStr := range rateLimitsMap {
		if duration, err := time.ParseDuration(limitStr); err == nil {
			c.RateLimits[gateway] = duration
		}
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
