package quotes

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"quotetrack/backend/internal/models"
)

// Store persists quotes and quote items in PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

const quoteCols = `q.id, q.user_id, q.customer_id, q.public_id, q.quote_number, q.title,
	q.notes, q.status,
	to_char(q.quote_date, 'YYYY-MM-DD'),
	CASE WHEN q.expiry_date IS NULL THEN NULL ELSE to_char(q.expiry_date, 'YYYY-MM-DD') END,
	CASE WHEN q.follow_up_date IS NULL THEN NULL ELSE to_char(q.follow_up_date, 'YYYY-MM-DD') END,
	q.total, q.created_at, q.updated_at`

// quoteReturnCols is quoteCols without the table alias, safe for
// INSERT ... RETURNING and UPDATE ... RETURNING.
const quoteReturnCols = `id, user_id, customer_id, public_id, quote_number, title,
	notes, status,
	to_char(quote_date, 'YYYY-MM-DD'),
	CASE WHEN expiry_date IS NULL THEN NULL ELSE to_char(expiry_date, 'YYYY-MM-DD') END,
	CASE WHEN follow_up_date IS NULL THEN NULL ELSE to_char(follow_up_date, 'YYYY-MM-DD') END,
	total, created_at, updated_at`

// ListQuery is a filter for listing quotes.
type ListQuery struct {
	Status     string
	CustomerID string
	Search     string
	// FollowUpDue true restricts to quotes needing follow-up today.
	FollowUpDue bool
	Today       string
	// Limit caps results (dashboard use).
	Limit int
}

// QuoteRow is a listed quote with its customer summary and line count.
type QuoteRow struct {
	models.Quote
	ItemCount int `json:"item_count"`
}

// List returns quotes for a user matching the filters, newest first.
func (s *Store) List(ctx context.Context, userID string, lq ListQuery) ([]QuoteRow, error) {
	query := `
		SELECT ` + quoteCols + `,
			c.id, c.name, c.phone, c.company,
			(SELECT COUNT(*)::int FROM quote_items qi WHERE qi.quote_id = q.id)
		FROM quotes q
		JOIN customers c ON c.id = q.customer_id
		WHERE q.user_id = $1`
	args := []interface{}{userID}
	n := 2

	if lq.Status != "" {
		query += fmt.Sprintf(` AND q.status = $%d`, n)
		args = append(args, lq.Status)
		n++
	}
	if lq.CustomerID != "" {
		query += fmt.Sprintf(` AND q.customer_id = $%d`, n)
		args = append(args, lq.CustomerID)
		n++
	}
	if lq.Search != "" {
		like := "%" + strings.ToLower(lq.Search) + "%"
		query += fmt.Sprintf(` AND (lower(q.title) LIKE $%d OR lower(q.quote_number) LIKE $%d)`, n, n)
		args = append(args, like)
		n++
	}
	if lq.FollowUpDue {
		query += fmt.Sprintf(` AND q.status NOT IN ('won','lost') AND q.follow_up_date IS NOT NULL AND q.follow_up_date <= $%d::date`, n)
		args = append(args, lq.Today)
		n++
	}

	query += ` ORDER BY q.created_at DESC`
	if lq.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, n)
		args = append(args, lq.Limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list quotes: %w", err)
	}
	defer rows.Close()

	list := []QuoteRow{}
	for rows.Next() {
		var q QuoteRow
		var custID, custName, custPhone, custCompany string
		err := rows.Scan(&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
			&q.Notes, &q.Status,
			&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
			&q.Total, &q.CreatedAt, &q.UpdatedAt,
			&custID, &custName, &custPhone, &custCompany,
			&q.ItemCount)
		if err != nil {
			return nil, fmt.Errorf("scan quote row: %w", err)
		}
		q.Customer = &models.Customer{ID: custID, Name: custName, Phone: custPhone, Company: custCompany}
		q.Subtotal = q.Total
		list = append(list, q)
	}
	return list, rows.Err()
}

// Get returns a single quote (with items) scoped to a user.
func (s *Store) Get(ctx context.Context, userID, id string) (*models.Quote, error) {
	var q models.Quote
	var custID, custName, custPhone, custEmail, custCompany string
	err := s.pool.QueryRow(ctx, `
		SELECT `+quoteCols+`, c.id, c.name, c.phone, c.email, c.company
		FROM quotes q
		JOIN customers c ON c.id = q.customer_id
		WHERE q.user_id = $1 AND q.id = $2`,
		userID, id,
	).Scan(&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
		&q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
		&q.Total, &q.CreatedAt, &q.UpdatedAt,
		&custID, &custName, &custPhone, &custEmail, &custCompany)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get quote: %w", err)
	}
	q.Customer = &models.Customer{ID: custID, Name: custName, Phone: custPhone, Email: custEmail, Company: custCompany}
	q.Subtotal = q.Total

	items, err := s.getItems(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	q.Items = items
	return &q, nil
}

// GetByPublicID returns a quote by its public share identifier and the
// owning business name. No user scoping: this is the unauthenticated path.
func (s *Store) GetByPublicID(ctx context.Context, publicID string) (*models.Quote, string, error) {
	var q models.Quote
	var custName string
	var businessName string
	err := s.pool.QueryRow(ctx, `
		SELECT q.id, q.customer_id, q.public_id, q.quote_number, q.title, q.notes, q.status,
		       to_char(q.quote_date, 'YYYY-MM-DD'),
		       CASE WHEN q.expiry_date IS NULL THEN NULL ELSE to_char(q.expiry_date, 'YYYY-MM-DD') END,
		       CASE WHEN q.follow_up_date IS NULL THEN NULL ELSE to_char(q.follow_up_date, 'YYYY-MM-DD') END,
		       q.total, c.name, u.business_name
		FROM quotes q
		JOIN customers c ON c.id = q.customer_id
		JOIN users u ON u.id = q.user_id
		WHERE q.public_id = $1`,
		publicID,
	).Scan(&q.ID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title, &q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate, &q.Total,
		&custName, &businessName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", pgx.ErrNoRows
		}
		return nil, "", fmt.Errorf("get public quote: %w", err)
	}
	q.Subtotal = q.Total
	q.Customer = &models.Customer{Name: custName}

	items, err := s.getItems(ctx, q.ID)
	if err != nil {
		return nil, "", err
	}
	q.Items = items
	return &q, businessName, nil
}

func (s *Store) getItems(ctx context.Context, quoteID string) ([]models.QuoteItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, position, description, quantity::float8, unit_price::bigint, total::bigint
		FROM quote_items WHERE quote_id = $1 ORDER BY position, created_at`,
		quoteID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	items := []models.QuoteItem{}
	for rows.Next() {
		var it models.QuoteItem
		if err := rows.Scan(&it.ID, &it.Position, &it.Description, &it.Quantity, &it.UnitPrice, &it.Total); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// CreateQuoteInput carries everything needed to create a quote.
type CreateQuoteInput struct {
	UserID       string
	CustomerID   string
	PublicID     string
	QuoteNumber  string
	Title        string
	Notes        string
	Status       models.QuoteStatus
	QuoteDate    string
	ExpiryDate   *string
	FollowUpDate *string
	Items        []models.QuoteItem
}

// Create inserts a quote with items in a transaction and returns it.
func (s *Store) Create(ctx context.Context, in *CreateQuoteInput) (*models.Quote, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify the customer belongs to the user.
	var custID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM customers WHERE id = $1 AND user_id = $2`,
		in.CustomerID, in.UserID).Scan(&custID)
	if err != nil {
		return nil, err // ErrNoRows propagated
	}

	// Sequential per-user quote number (e.g. 0001, 0002, ...).
	var seq int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM quotes WHERE user_id = $1`, in.UserID).Scan(&seq); err != nil {
		return nil, fmt.Errorf("count quotes: %w", err)
	}
	in.QuoteNumber = QuoteNumber(seq + 1)

	subtotal, _ := Totals(in.Items)

	var q models.Quote
	err = tx.QueryRow(ctx, `
		INSERT INTO quotes (user_id, customer_id, public_id, quote_number, title, notes, status,
		                   quote_date, expiry_date, follow_up_date, subtotal, total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, $9::date, $10::date, $11, $12)
		RETURNING `+quoteReturnCols,
		in.UserID, in.CustomerID, in.PublicID, in.QuoteNumber, in.Title, in.Notes, in.Status,
		in.QuoteDate, toNullableDate(in.ExpiryDate), toNullableDate(in.FollowUpDate),
		int64(subtotal), int64(subtotal),
	).Scan(&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
		&q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
		&q.Total, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert quote: %w", err)
	}
	q.Subtotal = q.Total

	if err := replaceItems(ctx, tx, q.ID, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit quote: %w", err)
	}

	items, err := s.getItems(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	q.Items = items
	c, err := s.customerSummary(ctx, q.CustomerID)
	if err == nil {
		q.Customer = c
	}
	return &q, nil
}

// Update replaces a quote's fields and items. The customer may change.
func (s *Store) Update(ctx context.Context, userID, id string, in *CreateQuoteInput) (*models.Quote, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Ensure the quote exists and belongs to the user; verify customer ownership.
	var currentCustomer string
	err = tx.QueryRow(ctx,
		`SELECT customer_id FROM quotes WHERE id = $1 AND user_id = $2`, id, userID).Scan(&currentCustomer)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx,
		`SELECT id FROM customers WHERE id = $1 AND user_id = $2`, in.CustomerID, userID).Scan(&in.CustomerID)
	if err != nil {
		return nil, err
	}

	subtotal, _ := Totals(in.Items)

	var q models.Quote
	err = tx.QueryRow(ctx, `
		UPDATE quotes SET
			customer_id = $2, title = $3, notes = $4, status = $5,
			quote_date = $6::date, expiry_date = $7::date, follow_up_date = $8::date,
			subtotal = $9, total = $9, updated_at = now()
		WHERE user_id = $1 AND id = $10
		RETURNING `+quoteReturnCols,
		userID, in.CustomerID, in.Title, in.Notes, in.Status,
		in.QuoteDate, toNullableDate(in.ExpiryDate), toNullableDate(in.FollowUpDate),
		int64(subtotal), id,
	).Scan(&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
		&q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
		&q.Total, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update quote: %w", err)
	}
	q.Subtotal = q.Total

	if err := replaceItems(ctx, tx, q.ID, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update: %w", err)
	}

	items, err := s.getItems(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	q.Items = items
	c, err := s.customerSummary(ctx, q.CustomerID)
	if err == nil {
		q.Customer = c
	}
	return &q, nil
}

func (s *Store) customerSummary(ctx context.Context, customerID string) (*models.Customer, error) {
	var c models.Customer
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, phone, email, company, notes, created_at, updated_at
		 FROM customers WHERE id = $1`, customerID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Company, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func replaceItems(ctx context.Context, tx pgx.Tx, quoteID string, items []models.QuoteItem) error {
	if _, err := tx.Exec(ctx, `DELETE FROM quote_items WHERE quote_id = $1`, quoteID); err != nil {
		return fmt.Errorf("clear items: %w", err)
	}
	for i, it := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO quote_items (quote_id, position, description, quantity, unit_price, total)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			quoteID, i, it.Description, it.Quantity, int64(it.UnitPrice), int64(it.Total)); err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}
	return nil
}

// SetStatus updates a quote status. When marked won/lost the follow-up date
// is cleared so the quote drops out of the active follow-up list.
func (s *Store) SetStatus(ctx context.Context, userID, id string, status models.QuoteStatus) (*models.Quote, error) {
	var q models.Quote
	var sqlQuery string
	if status == models.StatusWon || status == models.StatusLost {
		sqlQuery = `
			UPDATE quotes SET status = $3, follow_up_date = NULL, updated_at = now()
			WHERE user_id = $1 AND id = $2
			RETURNING ` + quoteReturnCols
	} else {
		sqlQuery = `
			UPDATE quotes SET status = $3, updated_at = now()
			WHERE user_id = $1 AND id = $2
			RETURNING ` + quoteReturnCols
	}

	err := s.pool.QueryRow(ctx, sqlQuery, userID, id, status).Scan(
		&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
		&q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
		&q.Total, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("set status: %w", err)
	}
	q.Subtotal = q.Total
	return &q, nil
}

// SetFollowUp updates the follow-up date on a quote.
func (s *Store) SetFollowUp(ctx context.Context, userID, id string, date *string) (*models.Quote, error) {
	var q models.Quote
	err := s.pool.QueryRow(ctx, `
		UPDATE quotes SET follow_up_date = $3::date, updated_at = now()
		WHERE user_id = $1 AND id = $2
		RETURNING `+quoteReturnCols,
		userID, id, toNullableDate(date),
	).Scan(&q.ID, &q.UserID, &q.CustomerID, &q.PublicID, &q.QuoteNumber, &q.Title,
		&q.Notes, &q.Status,
		&q.QuoteDate, &q.ExpiryDate, &q.FollowUpDate,
		&q.Total, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("set follow-up: %w", err)
	}
	q.Subtotal = q.Total
	return &q, nil
}

// Delete removes a quote scoped to the user (items cascade).
func (s *Store) Delete(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM quotes WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("delete quote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func toNullableDate(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}