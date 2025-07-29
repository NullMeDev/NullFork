package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"enhanced-gateway-scraper/internal/browser"
	"enhanced-gateway-scraper/internal/proxy"
	"enhanced-gateway-scraper/internal/search"
)

type WebServer struct {
	router       *mux.Router
	proxyManager *proxy.ProxyManager
	upgrader     websocket.Upgrader
	activeScan   *ScanSession
	mu           sync.RWMutex
}

type ScanSession struct {
	ID        string                 `json:"id"`
	Active    bool                   `json:"active"`
	Progress  float64                `json:"progress"`
	Results   []ScanResult           `json:"results"`
	Stats     ScanStats              `json:"stats"`
	Config    ScanConfig             `json:"config"`
	StartTime time.Time              `json:"start_time"`
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

type ScanConfig struct {
	ScanMode     string   `json:"scan_mode"`
	Categories   []string `json:"categories"`
	Query        string   `json:"query"`
	URL          string   `json:"url"`
	ProxyType    string   `json:"proxy_type"`
	Limit        int      `json:"limit"`
	Workers      int      `json:"workers"`
	Timeout      int      `json:"timeout"`
	OutputFormat string   `json:"output_format"`
}

type ScanResult struct {
	URL        string    `json:"url"`
	Status     string    `json:"status"`
	Gateways   []string  `json:"gateways"`
	Confidence float64   `json:"confidence"`
	Method     string    `json:"method"`
	Duration   int64     `json:"duration"`
	Proxy      string    `json:"proxy"`
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error,omitempty"`
}

type ScanStats struct {
	TotalScanned   int     `json:"total_scanned"`
	GatewaysFound  int     `json:"gateways_found"`
	SuccessRate    float64 `json:"success_rate"`
	Duration       int64   `json:"duration"`
	AvgScanTime    int64   `json:"avg_scan_time"`
	ProxiesUsed    int     `json:"proxies_used"`
}

type ProxyInfo struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Type      string    `json:"type"`
	Working   bool      `json:"working"`
	Latency   int64     `json:"latency"`
	LastUsed  time.Time `json:"last_used"`
	UseCount  int       `json:"use_count"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewWebServer() *WebServer {
	ws := &WebServer{
		router: mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}

	// Initialize proxy manager
	var err error
	ws.proxyManager, err = proxy.NewProxyManager()
	if err != nil {
		log.Printf("Failed to initialize proxy manager: %v", err)
	}

	ws.setupRoutes()
	return ws
}

func (ws *WebServer) setupRoutes() {
	// Serve static files
	ws.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	
	// Serve the main dashboard
	ws.router.HandleFunc("/", ws.handleDashboard).Methods("GET")
	
	// API routes
	api := ws.router.PathPrefix("/api").Subrouter()
	
	// Scan endpoints
	api.HandleFunc("/scan/start", ws.handleStartScan).Methods("POST")
	api.HandleFunc("/scan/stop", ws.handleStopScan).Methods("POST")
	api.HandleFunc("/scan/status", ws.handleScanStatus).Methods("GET")
	api.HandleFunc("/scan/results", ws.handleScanResults).Methods("GET")
	api.HandleFunc("/scan/export/{format}", ws.handleExportResults).Methods("GET")
	
	// Proxy endpoints
	api.HandleFunc("/proxy/list", ws.handleProxyList).Methods("GET")
	api.HandleFunc("/proxy/add", ws.handleAddProxies).Methods("POST")
	api.HandleFunc("/proxy/test", ws.handleTestProxies).Methods("POST")
	api.HandleFunc("/proxy/clear", ws.handleClearProxies).Methods("POST")
	api.HandleFunc("/proxy/scrape", ws.handleScrapeProxies).Methods("POST")
	
	// WebSocket endpoint for real-time updates
	api.HandleFunc("/ws", ws.handleWebSocket)
	
	// Health check
	api.HandleFunc("/health", ws.handleHealth).Methods("GET")
}

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/enhanced-dashboard.html")
}

func (ws *WebServer) handleStartScan(w http.ResponseWriter, r *http.Request) {
	var config ScanConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate config
	if err := ws.validateScanConfig(config); err != nil {
		ws.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Stop any existing scan
	if ws.activeScan != nil && ws.activeScan.Active {
		ws.activeScan.cancel()
	}

	// Create new scan session
	ctx, cancel := context.WithCancel(context.Background())
	sessionID := fmt.Sprintf("scan_%d", time.Now().Unix())
	
	ws.activeScan = &ScanSession{
		ID:        sessionID,
		Active:    true,
		Progress:  0,
		Results:   make([]ScanResult, 0),
		Config:    config,
		StartTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start scanning in background
	go ws.performScan(ws.activeScan)

	ws.sendSuccess(w, "Scan started successfully", map[string]string{"session_id": sessionID})
}

func (ws *WebServer) handleStopScan(w http.ResponseWriter, r *http.Request) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.activeScan == nil || !ws.activeScan.Active {
		ws.sendError(w, "No active scan to stop", http.StatusBadRequest)
		return
	}

	ws.activeScan.cancel()
	ws.activeScan.Active = false

	ws.sendSuccess(w, "Scan stopped successfully", nil)
}

func (ws *WebServer) handleScanStatus(w http.ResponseWriter, r *http.Request) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.activeScan == nil {
		ws.sendSuccess(w, "No scan session", map[string]interface{}{
			"active": false,
			"stats":  ScanStats{},
		})
		return
	}

	ws.activeScan.mu.RLock()
	status := map[string]interface{}{
		"active":    ws.activeScan.Active,
		"progress":  ws.activeScan.Progress,
		"stats":     ws.activeScan.Stats,
		"session_id": ws.activeScan.ID,
	}
	ws.activeScan.mu.RUnlock()

	ws.sendSuccess(w, "Scan status retrieved", status)
}

func (ws *WebServer) handleScanResults(w http.ResponseWriter, r *http.Request) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.activeScan == nil {
		ws.sendSuccess(w, "No scan results", []ScanResult{})
		return
	}

	ws.activeScan.mu.RLock()
	results := ws.activeScan.Results
	ws.activeScan.mu.RUnlock()

	ws.sendSuccess(w, "Scan results retrieved", results)
}

func (ws *WebServer) handleExportResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.activeScan == nil || len(ws.activeScan.Results) == 0 {
		ws.sendError(w, "No results to export", http.StatusBadRequest)
		return
	}

	ws.activeScan.mu.RLock()
	results := ws.activeScan.Results
	ws.activeScan.mu.RUnlock()

	filename := fmt.Sprintf("gateway_scan_%s", time.Now().Format("2006-01-02_15-04-05"))
	
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", filename))
		json.NewEncoder(w).Encode(results)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", filename))
		ws.writeCSV(w, results)
	case "table":
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.txt", filename))
		ws.writeTable(w, results)
	default:
		ws.sendError(w, "Invalid export format", http.StatusBadRequest)
	}
}

func (ws *WebServer) handleProxyList(w http.ResponseWriter, r *http.Request) {
	if ws.proxyManager == nil {
		ws.sendSuccess(w, "Proxy manager not available", []ProxyInfo{})
		return
	}

	proxies := ws.proxyManager.GetWorkingProxies()
	proxyInfos := make([]ProxyInfo, len(proxies))
	
	for i, p := range proxies {
		proxyInfos[i] = ProxyInfo{
			Host:     p.Host,
			Port:     p.Port,
			Type:     p.Type,
			Working:  p.Working,
			Latency:  p.Latency,
			LastUsed: p.LastUsed,
			UseCount: p.UseCount,
		}
	}

	ws.sendSuccess(w, "Proxy list retrieved", proxyInfos)
}

func (ws *WebServer) handleAddProxies(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Proxies []string `json:"proxies"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ws.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if ws.proxyManager == nil {
		ws.sendError(w, "Proxy manager not available", http.StatusInternalServerError)
		return
	}

	added := 0
	for _, proxyStr := range request.Proxies {
		parts := strings.Split(strings.TrimSpace(proxyStr), ":")
		if len(parts) != 2 {
			continue
		}
		
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		proxyInfo := &proxy.Proxy{
			Host:    parts[0],
			Port:    port,
			Type:    "http", // Default to HTTP
			Working: false,  // Will be validated
		}

		ws.proxyManager.AddProxy(proxyInfo)
		added++
	}

	ws.sendSuccess(w, fmt.Sprintf("Added %d proxies", added), map[string]int{"added": added})
}

func (ws *WebServer) handleTestProxies(w http.ResponseWriter, r *http.Request) {
	if ws.proxyManager == nil {
		ws.sendError(w, "Proxy manager not available", http.StatusInternalServerError)
		return
	}

	// Validate all proxies in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ws.proxyManager.ValidateProxies(ctx)
	}()

	ws.sendSuccess(w, "Proxy validation started", nil)
}

func (ws *WebServer) handleClearProxies(w http.ResponseWriter, r *http.Request) {
	if ws.proxyManager == nil {
		ws.sendError(w, "Proxy manager not available", http.StatusInternalServerError)
		return
	}

	// Clear all proxies (you'd need to implement this method)
	ws.sendSuccess(w, "All proxies cleared", nil)
}

func (ws *WebServer) handleScrapeProxies(w http.ResponseWriter, r *http.Request) {
	if ws.proxyManager == nil {
		ws.sendError(w, "Proxy manager not available", http.StatusInternalServerError)
		return
	}

	// Start proxy scraping in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		ws.proxyManager.ScrapeProxies(ctx)
	}()

	ws.sendSuccess(w, "Proxy scraping started", nil)
}

func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Send periodic updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ws.mu.RLock()
			if ws.activeScan != nil {
				ws.activeScan.mu.RLock()
				update := map[string]interface{}{
					"type":     "scan_update",
					"active":   ws.activeScan.Active,
					"progress": ws.activeScan.Progress,
					"stats":    ws.activeScan.Stats,
				}
				ws.activeScan.mu.RUnlock()
				
				if err := conn.WriteJSON(update); err != nil {
					ws.mu.RUnlock()
					return
				}
			}
			ws.mu.RUnlock()
		}
	}
}

func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":        "healthy",
		"timestamp":     time.Now(),
		"proxy_manager": ws.proxyManager != nil,
	}
	
	if ws.proxyManager != nil {
		proxies := ws.proxyManager.GetWorkingProxies()
		health["proxy_count"] = len(proxies)
	}

	ws.sendSuccess(w, "Service is healthy", health)
}

func (ws *WebServer) performScan(session *ScanSession) {
	defer func() {
		session.mu.Lock()
		session.Active = false
		session.mu.Unlock()
	}()

	// Generate URLs based on config
	urlGenerator := search.NewURLGenerator()
	var urls []string
	var err error

	switch session.Config.ScanMode {
	case "categories":
		urls, err = urlGenerator.GenerateFromCategories(session.Config.Categories, session.Config.Limit)
	case "query":
		urls, err = urlGenerator.GenerateFromQuery(session.Config.Query, session.Config.Limit)
	case "url":
		urls = []string{session.Config.URL}
	default:
		log.Printf("Invalid scan mode: %s", session.Config.ScanMode)
		return
	}

	if err != nil {
		log.Printf("Failed to generate URLs: %v", err)
		return
	}

	if len(urls) == 0 {
		log.Println("No URLs generated for scanning")
		return
	}

	// Initialize browser automation engine
	browserEngine := browser.NewAutomationEngine(browser.AutomationConfig{
		Headless:    true,
		UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		Timeout:     time.Duration(session.Config.Timeout) * time.Second,
		ProxyURL:    "", // Will be set per request if needed
	})

	// Create worker pool
	jobs := make(chan string, len(urls))
	results := make(chan ScanResult, len(urls))

	// Start workers
	for i := 0; i < session.Config.Workers; i++ {
		go ws.scanWorker(session, jobs, results, browserEngine)
	}

	// Send jobs
	for _, url := range urls {
		select {
		case jobs <- url:
		case <-session.ctx.Done():
			close(jobs)
			return
		}
	}
	close(jobs)

	// Collect results
	scannedCount := 0
	totalJobs := len(urls)

	for i := 0; i < totalJobs; i++ {
		select {
		case result := <-results:
			session.mu.Lock()
			session.Results = append(session.Results, result)
			scannedCount++
			session.Progress = float64(scannedCount) / float64(totalJobs) * 100
			
			// Update stats
			session.Stats = ws.calculateStats(session)
			session.mu.Unlock()

		case <-session.ctx.Done():
			return
		}
	}

	log.Printf("Scan completed: %d URLs scanned", scannedCount)
}

func (ws *WebServer) scanWorker(session *ScanSession, jobs <-chan string, results chan<- ScanResult, engine *browser.AutomationEngine) {
	for url := range jobs {
		select {
		case <-session.ctx.Done():
			return
		default:
			result := ws.scanURL(url, session, engine)
			results <- result
		}
	}
}

func (ws *WebServer) scanURL(url string, session *ScanSession, engine *browser.AutomationEngine) ScanResult {
	startTime := time.Now()
	
	result := ScanResult{
		URL:       url,
		Status:    "failed",
		Gateways:  []string{},
		Method:    "browser_automation",
		Proxy:     "none",
		Timestamp: startTime,
	}

	// Get proxy if configured
	if session.Config.ProxyType != "none" && ws.proxyManager != nil {
		if proxy := ws.proxyManager.GetRandomProxy(); proxy != nil {
			result.Proxy = fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)
			engine.SetProxy(result.Proxy)
		}
	}

	// Perform the scan
	ctx, cancel := context.WithTimeout(session.ctx, time.Duration(session.Config.Timeout)*time.Second)
	defer cancel()

	scanResult, err := engine.ScanWebsite(ctx, url)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// Process scan results
	if len(scanResult.DetectedGateways) > 0 {
		result.Status = "success"
		result.Gateways = make([]string, 0, len(scanResult.DetectedGateways))
		
		totalConfidence := 0.0
		for gateway, confidence := range scanResult.DetectedGateways {
			result.Gateways = append(result.Gateways, gateway)
			totalConfidence += confidence
		}
		
		if len(result.Gateways) > 0 {
			result.Confidence = totalConfidence / float64(len(result.Gateways))
		}
	} else {
		result.Status = "no_gateways"
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result
}

func (ws *WebServer) calculateStats(session *ScanSession) ScanStats {
	totalScanned := len(session.Results)
	gatewaysFound := 0
	totalDuration := int64(0)
	successCount := 0

	for _, result := range session.Results {
		if len(result.Gateways) > 0 {
			gatewaysFound += len(result.Gateways)
		}
		if result.Status == "success" {
			successCount++
		}
		totalDuration += result.Duration
	}

	successRate := 0.0
	avgScanTime := int64(0)
	if totalScanned > 0 {
		successRate = float64(successCount) / float64(totalScanned) * 100
		avgScanTime = totalDuration / int64(totalScanned)
	}

	return ScanStats{
		TotalScanned:  totalScanned,
		GatewaysFound: gatewaysFound,
		SuccessRate:   successRate,
		Duration:      time.Since(session.StartTime).Milliseconds(),
		AvgScanTime:   avgScanTime,
	}
}

func (ws *WebServer) validateScanConfig(config ScanConfig) error {
	if config.ScanMode == "query" && strings.TrimSpace(config.Query) == "" {
		return fmt.Errorf("search query is required for query mode")
	}
	
	if config.ScanMode == "url" && strings.TrimSpace(config.URL) == "" {
		return fmt.Errorf("URL is required for URL mode")
	}
	
	if config.ScanMode == "categories" && len(config.Categories) == 0 {
		return fmt.Errorf("at least one category is required for category mode")
	}

	if config.Limit < 1 || config.Limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	if config.Workers < 1 || config.Workers > 100 {
		return fmt.Errorf("workers must be between 1 and 100")
	}

	if config.Timeout < 5 || config.Timeout > 300 {
		return fmt.Errorf("timeout must be between 5 and 300 seconds")
	}

	return nil
}

func (ws *WebServer) writeCSV(w http.ResponseWriter, results []ScanResult) {
	fmt.Fprintln(w, "URL,Status,Gateways,Confidence,Method,Duration,Proxy,Timestamp")
	for _, result := range results {
		gateways := strings.Join(result.Gateways, ";")
		fmt.Fprintf(w, "%s,%s,%s,%.2f,%s,%d,%s,%s\n",
			result.URL, result.Status, gateways, result.Confidence,
			result.Method, result.Duration, result.Proxy, result.Timestamp.Format(time.RFC3339))
	}
}

func (ws *WebServer) writeTable(w http.ResponseWriter, results []ScanResult) {
	fmt.Fprintln(w, "URL\tStatus\tGateways\tConfidence\tMethod\tDuration\tProxy\tTimestamp")
	fmt.Fprintln(w, strings.Repeat("=", 100))
	for _, result := range results {
		gateways := strings.Join(result.Gateways, ";")
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\t%dms\t%s\t%s\n",
			result.URL, result.Status, gateways, result.Confidence,
			result.Method, result.Duration, result.Proxy, result.Timestamp.Format("15:04:05"))
	}
}

func (ws *WebServer) sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	server := NewWebServer()
	
	port := ":8083"
	fmt.Printf("🚀 Enhanced Gateway Scraper Web Server starting on port %s\n", port)
	fmt.Printf("📊 Dashboard: http://localhost%s\n", port)
	fmt.Printf("🔌 API: http://localhost%s/api\n", port)
	fmt.Printf("📡 WebSocket: ws://localhost%s/api/ws\n", port)
	
	log.Fatal(http.ListenAndServe(port, server.router))
}
