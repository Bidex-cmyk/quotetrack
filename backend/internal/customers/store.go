package customers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/models"
)

// Store persists customers in PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const customerCols = `c.id, c.user_id, c.name, c.phone, c.email, c.company, c.notes, c.created_at, c.updated_at`

// customerReturnCols is customerCols without the table alias, safe for
// INSERT/UPDATE ... RETURNING.
const customerReturnCols = `id, user_id, name, phone, email, company, notes, created_at, updated_at`

func scanCustomer(row pgx.Row) (*models.Customer, error) {
	var c models.Customer
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan customer: %w", err)
	}
	return &c, nil
}

// List returns all customers for a user, optionally filtered by a search term.
func (s *Store) List(ctx context.Context, userID, search string) ([]models.CustomerWithStats, error) {
	query := `
		SELECT ` + customerCols + `,
		       COUNT(q.id)::int AS quote_count,
		       COALESCE(SUM(q.total), 0)::bigint AS total_value
		FROM customers c
		LEFT JOIN quotes q ON q.customer_id = c.id
		WHERE c.user_id = $1`
	args := []interface{}{userID}

	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query += ` AND (lower(c.name) LIKE $2 OR lower(c.company) LIKE $2 OR lower(c.email) LIKE $2 OR c.phone LIKE $2)`
		args = append(args, like)
	}

	query += ` GROUP BY c.id ORDER BY c.name`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	list := []models.CustomerWithStats{}
	for rows.Next() {
		var c models.CustomerWithStats
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt, &c.QuoteCount, &c.TotalValue)
		if err != nil {
			return nil, fmt.Errorf("scan customer row: %w", err)
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// Get returns a single customer scoped to a user. Returns a
// pgx.ErrNoRows-equivalent error if not found.
func (s *Store) Get(ctx context.Context, userID, id string) (*models.CustomerWithStats, error) {
	var c models.CustomerWithStats
	err := s.pool.QueryRow(ctx, `
		SELECT `+customerCols+`,
		       COUNT(q.id)::int,
		       COALESCE(SUM(q.total), 0)::bigint
		FROM customers c
		LEFT JOIN quotes q ON q.customer_id = c.id
		WHERE c.user_id = $1 AND c.id = $2
		GROUP BY c.id`,
		userID, id,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt, &c.QuoteCount, &c.TotalValue)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &c, nil
}

// Create inserts a new customer.
func (s *Store) Create(ctx context.Context, userID string, c *models.Customer) (*models.Customer, error) {
	c.UserID = userID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO customers (user_id, name, phone, email, company, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+customerReturnCols,
		userID, c.Name, c.Phone, c.Email, c.Company, c.Notes,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	return c, nil
}

// Update updates a customer row scoped to the user.
func (s *Store) Update(ctx context.Context, userID string, c *models.Customer) (*models.Customer, error) {
	err := s.pool.QueryRow(ctx, `
		UPDATE customers
		SET name = $3, phone = $4, email = $5, company = $6, notes = $7, updated_at = now()
		WHERE user_id = $1 AND id = $2
		RETURNING `+customerReturnCols,
		userID, c.ID, c.Name, c.Phone, c.Email, c.Company, c.Notes,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("update customer: %w", err)
	}
	return c, nil
}

// Delete removes a customer scoped to the user. Quotes cascade.
func (s *Store) Delete(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM customers WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}