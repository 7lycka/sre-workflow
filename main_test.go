package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHealthHandler はヘルスチェックエンドポイントのテスト
// SREワークフローの重要要素：ヘルスチェックの動作保証
func TestHealthHandler(t *testing.T) {
	// テスト用のHTTPリクエスト作成
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// レスポンス記録用のRecorder作成
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	// ハンドラー実行
	handler.ServeHTTP(rr, req)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Content-Type確認
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v",
			ct, expected)
	}

	// JSONレスポンス構造確認
	var health HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	// レスポンスフィールド確認
	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}

	if health.Version == "" {
		t.Error("Expected version to be set")
	}

	// タイムスタンプ形式確認（RFC3339形式であることを確認）
	if _, err := time.Parse(time.RFC3339, health.Timestamp); err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

// TestMetricsHandler はメトリクスエンドポイントのテスト
// SRE監視要件：メトリクス取得機能の動作保証
func TestMetricsHandler(t *testing.T) {
	// 初期リクエストカウントを記録
	initialCount := requestCount

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(metricsHandler)

	handler.ServeHTTP(rr, req)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// JSONレスポンス構造確認
	var metrics MetricsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &metrics); err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	// リクエストカウントが増加していることを確認
	if metrics.RequestCount <= initialCount {
		t.Errorf("Request count should increase: got %d, initial was %d",
			metrics.RequestCount, initialCount)
	}

	// アップタイムが正の値であることを確認
	if metrics.Uptime <= 0 {
		t.Errorf("Uptime should be positive: got %f", metrics.Uptime)
	}

	// メモリ使用量が設定されていることを確認
	if metrics.MemoryUsageMB <= 0 {
		t.Errorf("Memory usage should be positive: got %d", metrics.MemoryUsageMB)
	}
}

// TestRootHandler はルートエンドポイントのテスト
// 基本的なWebページ配信機能の確認
func TestRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rootHandler)

	handler.ServeHTTP(rr, req)

	// ステータスコード確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Content-Type確認
	expected := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v",
			ct, expected)
	}

	// HTML内容確認
	body := rr.Body.String()
	if !strings.Contains(body, "SRE Workflow Demo") {
		t.Error("Response should contain 'SRE Workflow Demo'")
	}

	if !strings.Contains(body, "/health") {
		t.Error("Response should contain link to health endpoint")
	}

	if !strings.Contains(body, "/metrics") {
		t.Error("Response should contain link to metrics endpoint")
	}
}

// TestLogMiddleware はログミドルウェアのテスト
// SREのログ出力要件確認（ログ出力が正常に動作することを保証）
func TestLogMiddleware(t *testing.T) {
	// テスト用のハンドラー
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}

	// ミドルウェア適用
	wrappedHandler := logMiddleware(testHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	wrappedHandler(rr, req)

	// ミドルウェアが正常に動作したことを確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Middleware affected status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if body != "test" {
		t.Errorf("Middleware affected response body: got %v want %v",
			body, "test")
	}
}

// BenchmarkHealthHandler はヘルスチェックエンドポイントのベンチマークテスト
// SREパフォーマンス要件：レスポンス時間の測定
func BenchmarkHealthHandler(b *testing.B) {
	req, _ := http.NewRequest("GET", "/health", nil)
	
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(healthHandler)
		handler.ServeHTTP(rr, req)
	}
}

// BenchmarkMetricsHandler はメトリクスエンドポイントのベンチマークテスト
// SREパフォーマンス要件：監視エンドポイントの応答性能測定
func BenchmarkMetricsHandler(b *testing.B) {
	req, _ := http.NewRequest("GET", "/metrics", nil)
	
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(metricsHandler)
		handler.ServeHTTP(rr, req)
	}
}