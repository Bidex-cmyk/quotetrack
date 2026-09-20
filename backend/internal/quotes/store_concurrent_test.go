package quotes

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/auth"
	"quotetrack/backend/internal/database"
	"quotetrack/backend/internal/models"
)

// TestConcurrentQuoteNumbering proves that concurrent quote creation cannot
// produce duplicate quote numbers for the same user, and that numbers stay
// sequential per user without being reused after a deletion.
//
// It needs a disposable PostgreSQL database in TEST_DATABASE_URL and is
// skipped otherwise, so `go test ./...` remains hermetic.
//
//	TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/quotetrack_test?sslmode=disable \
//	    go test ./internal/quotes/ -run TestConcurrentQuoteNumbering -v
func TestConcurrentQuoteNumbering(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed concurrency test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Isolate the test in its own schema so it never touches other data.
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
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_, _ = admin.Exec(cleanupCtx, fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schema))
	}()

	pool, err := pgxpool.New(ctx, dsnWithParam(dsn, "search_path", schema))
	if err != nil {
		t.Fatalf("connect pool: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Seed two users, one customer each.
	users := []string{uuid.NewString(), uuid.NewString()}
	for i, uid := range users {
		if _, err := pool.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, business_name) VALUES ($1, $2, 'x', $3)`,
			uid, fmt.Sprintf("concurrent-%d@test.local", i), fmt.Sprintf("Biz %d", i)); err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO customers (id, user_id, name) VALUES ($1, $2, 'Customer')`,
			uuid.NewString(), uid); err != nil {
			t.Fatalf("seed customer: %v", err)
		}
	}

	store := NewStore(pool)
	const goroutines, perGoroutine = 8, 5

	for _, uid := range users {
		var custID string
		if err := pool.QueryRow(ctx,
			`SELECT id FROM customers WHERE user_id = $1 LIMIT 1`, uid).Scan(&custID); err != nil {
			t.Fatalf("load customer: %v", err)
		}

		var (
			wg     sync.WaitGroup
			mu     sync.Mutex
			errs   []error
			number []string
		)
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					q, err := store.Create(ctx, &CreateQuoteInput{
						UserID:     uid,
						CustomerID: custID,
						PublicID:   mustPublicID(t),
						Title:      "Concurrent quote",
						Status:     models.StatusDraft,
						QuoteDate:  "2026-09-20",
						Items: []models.QuoteItem{
							{Description: "item", Quantity: 1, UnitPrice: 1000},
						},
					})
					mu.Lock()
					if err != nil {
						errs = append(errs, err)
					} else {
						number = append(number, q.QuoteNumber)
					}
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		total := goroutines * perGoroutine
		if len(errs) != 0 {
			t.Fatalf("concurrent creation returned %d errors, first: %v", len(errs), errs[0])
		}
		if len(number) != total {
			t.Fatalf("created %d quotes, want %d", len(number), total)
		}

		// Every number must be unique, sequential and gap-free for this user.
		want := map[string]bool{}
		for n := 1; n <= total; n++ {
			want[QuoteNumber(n)] = true
		}
		got := map[string]bool{}
		for _, n := range number {
			if got[n] {
				t.Errorf("duplicate quote number %s for user %s", n, uid)
			}
			got[n] = true
		}
		for n := range want {
			if !got[n] {
				t.Errorf("user %s: missing expected quote number %s", uid, n)
			}
		}
		for n := range got {
			if !want[n] {
				t.Errorf("user %s: unexpected quote number %s", uid, n)
			}
		}

		// A deleted number must never be reused (the old COUNT(*) generator
		// would hand out QT-0040 again here, which already exists).
		if _, err := pool.Exec(ctx,
			`DELETE FROM quotes WHERE user_id = $1 AND quote_number = $2`, uid, QuoteNumber(7)); err != nil {
			t.Fatalf("delete quote: %v", err)
		}
		q, err := store.Create(ctx, &CreateQuoteInput{
			UserID:     uid,
			CustomerID: custID,
			PublicID:   mustPublicID(t),
			Title:      "Post-delete quote",
			Status:     models.StatusDraft,
			QuoteDate:  "2026-09-20",
			Items: []models.QuoteItem{
				{Description: "item", Quantity: 1, UnitPrice: 1000},
			},
		})
		if err != nil {
			t.Fatalf("create after delete: %v", err)
		}
		if wantNum := QuoteNumber(total + 1); q.QuoteNumber != wantNum {
			t.Errorf("quote number after deletion = %s, want %s", q.QuoteNumber, wantNum)
		}
	}
}

// dsnWithParam appends a connection parameter whether or not the DSN already
// carries a query string.
func dsnWithParam(dsn, key, value string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + key + "=" + value
}

func mustPublicID(t *testing.T) string {
	t.Helper()
	id, err := auth.RandomPublicID()
	if err != nil {
		t.Fatalf("public id: %v", err)
	}
	return id
}
