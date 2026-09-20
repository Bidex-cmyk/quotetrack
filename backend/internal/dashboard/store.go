package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/models"
	"quotetrack/backend/internal/quotes"
)

// Store aggregates dashboard and analytics queries.
type Store struct {
	pool        *pgxpool.Pool
	quotesStore *quotes.Store
}

func NewStore(pool *pgxpool.Pool, quotesStore *quotes.Store) *Store {
	return &Store{pool: pool, quotesStore: quotesStore}
}

// Stats holds the headline figures shown on the dashboard.
type Stats struct {
	TotalQuotes  int `json:"total_quotes"`
	Waiting      int `json:"waiting"`
	FollowUpsDue int `json:"follow_ups_due_today"`
	Won          int `json:"won"`
	Lost         int `json:"lost"`
}

// GetStats computes counts for the current user.
func (s *Store) GetStats(ctx context.Context, userID string) (*Stats, error) {
	stats := &Stats{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int AS total_quotes,
			COUNT(*) FILTER (WHERE status = 'waiting')::int AS waiting,
			COUNT(*) FILTER (WHERE status = 'won')::int AS won,
			COUNT(*) FILTER (WHERE status = 'lost')::int AS lost,
			COUNT(*) FILTER (WHERE status NOT IN ('won','lost')
			                    AND follow_up_date IS NOT NULL
			                    AND follow_up_date <= CURRENT_DATE)::int AS follow_ups_due
		FROM quotes WHERE user_id = $1`, userID,
	).Scan(&stats.TotalQuotes, &stats.Waiting, &stats.Won, &stats.Lost, &stats.FollowUpsDue)
	if err != nil {
		return nil, fmt.Errorf("dashboard stats: %w", err)
	}
	return stats, nil
}

// FollowUpsDue returns quotes that need follow-up today (follow_up_date <= today,
// status not won/lost), newest relevant first.
func (s *Store) FollowUpsDue(ctx context.Context, userID string, today time.Time) ([]quotes.QuoteRow, error) {
	todayStr := today.Format("2006-01-02")
	lq := quotes.ListQuery{FollowUpDue: true, Today: todayStr, Limit: 50}
	return s.quotesStore.List(ctx, userID, lq)
}

// Analytics holds basic win/loss metrics.
type Analytics struct {
	TotalQuotes  int          `json:"total_quotes"`
	TotalValue   models.Cents `json:"total_value"`
	Won          int          `json:"won"`
	WonValue     models.Cents `json:"won_value"`
	Lost         int          `json:"lost"`
	LostValue    models.Cents `json:"lost_value"`
	Waiting      int          `json:"waiting"`
	WaitingValue models.Cents `json:"waiting_value"`
	WinRate      float64      `json:"win_rate"`
}

// GetAnalytics computes the analytics summary for a user.
func (s *Store) GetAnalytics(ctx context.Context, userID string) (*Analytics, error) {
	a := &Analytics{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int AS total_quotes,
			COALESCE(SUM(total), 0)::bigint AS total_value,
			COUNT(*) FILTER (WHERE status = 'won')::int AS won,
			COALESCE(SUM(total) FILTER (WHERE status = 'won'), 0)::bigint AS won_value,
			COUNT(*) FILTER (WHERE status = 'lost')::int AS lost,
			COALESCE(SUM(total) FILTER (WHERE status = 'lost'), 0)::bigint AS lost_value,
			COUNT(*) FILTER (WHERE status = 'waiting')::int AS waiting,
			COALESCE(SUM(total) FILTER (WHERE status = 'waiting'), 0)::bigint AS waiting_value
		FROM quotes WHERE user_id = $1`, userID,
	).Scan(&a.TotalQuotes, &a.TotalValue, &a.Won, &a.WonValue, &a.Lost, &a.LostValue, &a.Waiting, &a.WaitingValue)
	if err != nil {
		return nil, fmt.Errorf("analytics: %w", err)
	}
	decided := a.Won + a.Lost
	if decided > 0 {
		a.WinRate = float64(a.Won) / float64(decided)
	}
	return a, nil
}
