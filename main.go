package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// HealthResponse はヘルスチェックAPIのレスポンス構造体
// SREワークフローでの監視・ロードバランサーからの生存確認に使用
type HealthResponse struct {
	Status    string `json:"status"`    // サービス状態 ("healthy" など)
	Timestamp string `json:"timestamp"` // 現在時刻（RFC3339形式）
	Version   string `json:"version"`   // アプリケーションバージョン
}

// MetricsResponse はメトリクス取得APIのレスポンス構造体
// Prometheus形式での監視データ提供用
type MetricsResponse struct {
	RequestCount  int64   `json:"request_count"`   // 総リクエスト数
	Uptime        float64 `json:"uptime_seconds"`  // サービス稼働時間（秒）
	MemoryUsageMB int64   `json:"memory_usage_mb"` // メモリ使用量（MB）
}

// グローバル変数でアプリケーション開始時刻とリクエストカウンターを管理
var (
	startTime    = time.Now()
	requestCount int64
)

// healthHandler はヘルスチェックエンドポイント
// Kubernetes/Cloud Run のヘルスチェック、ロードバランサー監視で使用
// SREの可観測性（Observability）要件を満たす重要なエンドポイント
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// リクエストカウンターをインクリメント（実装簡略化）
	requestCount++

	// アプリケーションバージョンを環境変数から取得（デフォルト値設定）
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	// ヘルスチェックレスポンスを構築
	health := HealthResponse{
		Status:    "healthy",                       // 常に健康状態を返す（本格実装では内部状態をチェック）
		Timestamp: time.Now().Format(time.RFC3339), // RFC3339形式の現在時刻
		Version:   version,                         // アプリケーションバージョン
	}

	// JSONレスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// JSONエンコードしてレスポンス送信
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Health check accessed - Status: healthy, Version: %s", version)
}

// metricsHandler はメトリクス取得エンドポイント
// Prometheus監視システムやAPMツールでの性能監視に使用
// SREのSLI/SLO監視に必要なメトリクス提供
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	// サービス稼働時間を計算
	uptime := time.Since(startTime).Seconds()

	// メモリ使用量を簡易取得（実装簡略化）
	// 実際の本格実装では runtime.MemStats を使用
	var memStats int64 = 50 // MB単位での仮想値

	// メトリクスレスポンスを構築
	metrics := MetricsResponse{
		RequestCount:  requestCount,
		Uptime:        uptime,
		MemoryUsageMB: memStats,
	}

	// JSONレスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// JSONエンコードしてレスポンス送信
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Metrics accessed - Requests: %d, Uptime: %.2fs", requestCount, uptime)
}

// rootHandler はルートパスのハンドラー
// 基本的なサービス情報を提供するランディングページ
func rootHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	// シンプルなHTMLレスポンス
	html := `<!DOCTYPE html>
<html>
<head>
    <title>SRE Workflow Demo</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>SRE Workflow Demo Application</h1>
    <p>Golang製のSREワークフロー検証用アプリケーションです。</p>
    <ul>
        <li><a href="/health">Health Check</a> - サービス生存確認</li>
        <li><a href="/metrics">Metrics</a> - 監視用メトリクス</li>
    </ul>
    <p>Container Image: 署名付きでセキュアにデプロイ済み</p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, html)

	log.Printf("Root page accessed from %s", r.RemoteAddr)
}

// logMiddleware はHTTPリクエストをログ出力するミドルウェア
// SREの監視要件：すべてのリクエストをトレース可能にする
func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// リクエスト処理を実行
		next(w, r)

		// 処理時間とリクエスト情報をログ出力
		duration := time.Since(start)
		log.Printf("%s %s %s - Duration: %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			duration)
	}
}

func main() {
	// ポート番号を環境変数から取得（Cloud Run では PORT が自動設定される）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Golangの一般的なデフォルトポート
	}

	// アプリケーション開始ログ
	log.Printf("Starting SRE Workflow Demo Server on port %s", port)
	log.Printf("Start time: %s", startTime.Format(time.RFC3339))

	// HTTPルーティング設定
	// ミドルウェアを適用してすべてのリクエストをログ出力
	http.HandleFunc("/", logMiddleware(rootHandler))
	http.HandleFunc("/health", logMiddleware(healthHandler))
	http.HandleFunc("/metrics", logMiddleware(metricsHandler))

	// HTTPサーバー設定
	// 本格的なSREワークフローではタイムアウト設定が重要
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second, // リクエスト読み取りタイムアウト
		WriteTimeout: 15 * time.Second, // レスポンス書き込みタイムアウト
		IdleTimeout:  60 * time.Second, // アイドル接続タイムアウト
	}

	// HTTPサーバー開始
	log.Printf("Server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
