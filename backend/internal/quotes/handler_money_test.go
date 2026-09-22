package quotes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/config"
	"quotetrack/backend/internal/database"
	"quotetrack/backend/internal/server"
)

func TestCreateQuoteMoneyUnits(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed handler test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	schema := "qt_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA %s`, schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(context.Background(), fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schema))
	}()

	pool, err := pgxpool.New(ctx, dsnWithSearchPath(dsn, schema))
	if err != nil {
		t.Fatalf("connect pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var userID, customerID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, business_name) VALUES ('money@test.local', 'x', 'Money Test') RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO customers (user_id, name) VALUES ($1, 'Customer') RETURNING id`, userID).Scan(&customerID); err != nil {
		t.Fatalf("create customer: %v", err)
	}

	const secret = "money-test-secret"
	token, err := auth.NewTokenManager(secret, time.Hour).Create(userID)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	h := server.New(&config.Config{JWTSecret: secret, JWTExpiry: "1h"}, pool)
	body := fmt.Sprintf(`{"customer_id":%q,"title":"Money quote","status":"draft","quote_date":"2026-09-22","items":[{"description":"Service","quantity":1,"unit_price":50000}]}`, customerID)
	req := httptest.NewRequest(http.MethodPost, "/api/quotes", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var got struct {
		Total float64 `json:"total"`
		Items []struct {
			UnitPrice float64 `json:"unit_price"`
			Total     float64 `json:"total"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("response has %d items, want 1", len(got.Items))
	}
	if got.Items[0].UnitPrice != 50000 || got.Items[0].Total != 50000 || got.Total != 50000 {
		t.Errorf("response money = unit_price %v, item total %v, quote total %v; want 50000 each", got.Items[0].UnitPrice, got.Items[0].Total, got.Total)
	}

	var unitPrice int64
	if err := pool.QueryRow(ctx, `SELECT unit_price FROM quote_items LIMIT 1`).Scan(&unitPrice); err != nil {
		t.Fatalf("read persisted unit price: %v", err)
	}
	if unitPrice != 5000000 {
		t.Errorf("persisted unit_price = %d, want 5000000 cents", unitPrice)
	}
}

func dsnWithSearchPath(dsn, schema string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "search_path=" + schema
}
