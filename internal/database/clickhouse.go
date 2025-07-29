package database

import (
	"context"
	"fmt"

	"enhanced-gateway-scraper/pkg/types"
	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseClient struct {
	conn clickhouse.Conn
}

func NewClickHouseClient(dsn string) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: "gateway_scraper",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		return nil, err
	}

	client := &ClickHouseClient{conn: conn}
	if err := client.ensureTables(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ClickHouseClient) ensureTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS proxies (
			id String,
			host String,
			port UInt16,
			type String,
			username String,
			password String,
			country String,
			city String,
			isp String,
			anonymity String,
			working UInt8,
			latency UInt32,
			quality_score UInt8,
			last_test DateTime,
			fail_count UInt32,
			success_count UInt32,
			created_at DateTime,
			updated_at DateTime
		) ENGINE = MergeTree() ORDER BY (id, created_at)`,

		`CREATE TABLE IF NOT EXISTS gateways (
			id String,
			domain String,
			url String,
			gateway_type String,
			gateway_name String,
			confidence Float32,
			detection_method String,
			patterns Array(String),
			metadata Map(String, String),
			screenshot String,
			status_code UInt16,
			response_size UInt32,
			load_time UInt32,
			created_at DateTime,
			updated_at DateTime,
			last_checked DateTime
		) ENGINE = MergeTree() ORDER BY (domain, created_at)`,

		`CREATE TABLE IF NOT EXISTS check_results (
			id String,
			combo_username String,
			combo_password String,
			combo_email String,
			config_name String,
			status String,
			response String,
			error String,
			proxy_host String,
			proxy_port UInt16,
			latency UInt32,
			timestamp DateTime
		) ENGINE = MergeTree() ORDER BY (timestamp, config_name)`,

		`CREATE TABLE IF NOT EXISTS scan_sessions (
			id String,
			type String,
			status String,
			config Map(String, String),
			start_time DateTime,
			end_time Nullable(DateTime),
			error_message String,
			results_path String
		) ENGINE = MergeTree() ORDER BY (start_time, type)`,
	}

	for _, query := range queries {
		if err := c.conn.Exec(context.Background(), query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (c *ClickHouseClient) InsertProxy(proxy types.Proxy) error {
	query := `INSERT INTO proxies (id, host, port, type, username, password, country, city, isp, anonymity, working, latency, quality_score, last_test, fail_count, success_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	return c.conn.Exec(context.Background(), query,
		proxy.ID, proxy.Host, proxy.Port, string(proxy.Type), proxy.Username, proxy.Password,
		proxy.Country, proxy.City, proxy.ISP, proxy.Anonymity, proxy.Working, proxy.Latency,
		proxy.QualityScore, proxy.LastTest, proxy.FailCount, proxy.SuccessCount,
		proxy.CreatedAt, proxy.UpdatedAt)
}

func (c *ClickHouseClient) InsertGateway(gateway types.Gateway) error {
	query := `INSERT INTO gateways (id, domain, url, gateway_type, gateway_name, confidence, detection_method, patterns, metadata, screenshot, status_code, response_size, load_time, created_at, updated_at, last_checked) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	return c.conn.Exec(context.Background(), query,
		gateway.ID, gateway.Domain, gateway.URL, gateway.GatewayType, gateway.GatewayName,
		gateway.Confidence, gateway.DetectionMethod, gateway.Patterns, gateway.Metadata,
		gateway.Screenshot, gateway.StatusCode, gateway.ResponseSize, gateway.LoadTime,
		gateway.CreatedAt, gateway.UpdatedAt, gateway.LastChecked)
}

func (c *ClickHouseClient) InsertCheckResult(result types.CheckResult) error {
	query := `INSERT INTO check_results (id, combo_username, combo_password, combo_email, config_name, status, response, error, proxy_host, proxy_port, latency, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	var proxyHost string
	var proxyPort uint16
	if result.Proxy != nil {
		proxyHost = result.Proxy.Host
		proxyPort = uint16(result.Proxy.Port)
	}

	return c.conn.Exec(context.Background(), query,
		result.ID, result.Combo.Username, result.Combo.Password, result.Combo.Email,
		result.Config, result.Status, result.Response, result.Error,
		proxyHost, proxyPort, result.Latency, result.Timestamp)
}

func (c *ClickHouseClient) GetProxies(ctx context.Context, limit int) ([]types.Proxy, error) {
	query := `SELECT id, host, port, type, username, password, country, city, isp, anonymity, working, latency, quality_score, last_test, fail_count, success_count, created_at, updated_at FROM proxies WHERE working = 1 ORDER BY quality_score DESC LIMIT ?`
	
	rows, err := c.conn.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []types.Proxy
	for rows.Next() {
		var proxy types.Proxy
		var proxyType string
		err := rows.Scan(&proxy.ID, &proxy.Host, &proxy.Port, &proxyType, &proxy.Username, &proxy.Password,
			&proxy.Country, &proxy.City, &proxy.ISP, &proxy.Anonymity, &proxy.Working, &proxy.Latency,
			&proxy.QualityScore, &proxy.LastTest, &proxy.FailCount, &proxy.SuccessCount,
			&proxy.CreatedAt, &proxy.UpdatedAt)
		if err != nil {
			continue
		}
		proxy.Type = types.ProxyType(proxyType)
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func (c *ClickHouseClient) GetGateways(ctx context.Context, limit int) ([]types.Gateway, error) {
	query := `SELECT id, domain, url, gateway_type, gateway_name, confidence, detection_method, patterns, metadata, screenshot, status_code, response_size, load_time, created_at, updated_at, last_checked FROM gateways ORDER BY confidence DESC LIMIT ?`
	
	rows, err := c.conn.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gateways []types.Gateway
	for rows.Next() {
		var gateway types.Gateway
		err := rows.Scan(&gateway.ID, &gateway.Domain, &gateway.URL, &gateway.GatewayType, &gateway.GatewayName,
			&gateway.Confidence, &gateway.DetectionMethod, &gateway.Patterns, &gateway.Metadata,
			&gateway.Screenshot, &gateway.StatusCode, &gateway.ResponseSize, &gateway.LoadTime,
			&gateway.CreatedAt, &gateway.UpdatedAt, &gateway.LastChecked)
		if err != nil {
			continue
		}
		gateways = append(gateways, gateway)
	}

	return gateways, nil
}

// StoreProxy stores a proxy in the database
func (c *ClickHouseClient) StoreProxy(ctx context.Context, proxy types.Proxy) error {
	return c.InsertProxy(proxy)
}

// GetProxyCount returns the total number of proxies
func (c *ClickHouseClient) GetProxyCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM proxies`
	err := c.conn.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetWorkingProxyCount returns the number of working proxies
func (c *ClickHouseClient) GetWorkingProxyCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM proxies WHERE working = 1`
	err := c.conn.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetGatewayCount returns the total number of gateways
func (c *ClickHouseClient) GetGatewayCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM gateways`
	err := c.conn.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetGatewaysByDomain returns gateways for a specific domain
func (c *ClickHouseClient) GetGatewaysByDomain(ctx context.Context, domain string) ([]types.Gateway, error) {
	query := `SELECT id, domain, url, gateway_type, gateway_name, confidence, detection_method, patterns, metadata, screenshot, status_code, response_size, load_time, created_at, updated_at, last_checked FROM gateways WHERE domain = ? ORDER BY confidence DESC`
	
	rows, err := c.conn.Query(ctx, query, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gateways []types.Gateway
	for rows.Next() {
		var gateway types.Gateway
		err := rows.Scan(&gateway.ID, &gateway.Domain, &gateway.URL, &gateway.GatewayType, &gateway.GatewayName,
			&gateway.Confidence, &gateway.DetectionMethod, &gateway.Patterns, &gateway.Metadata,
			&gateway.Screenshot, &gateway.StatusCode, &gateway.ResponseSize, &gateway.LoadTime,
			&gateway.CreatedAt, &gateway.UpdatedAt, &gateway.LastChecked)
		if err != nil {
			continue
		}
		gateways = append(gateways, gateway)
	}

	return gateways, nil
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}
