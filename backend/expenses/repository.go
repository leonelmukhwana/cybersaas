package expenses

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("expense not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

type CreateExpenseParams struct {
	ID            string
	TenantID      string
	BranchID      string
	RecordedBy    string
	Category      string
	Description   *string
	Amount        string
	Currency      string
	PaymentMethod string
	Reference     *string
	ExpenseDate   interface{}
	Status        string
}

func (r *Repository) Create(
	ctx context.Context,
	params CreateExpenseParams,
) (*Expense, error) {
	const query = `
		INSERT INTO expenses (
			id,
			tenant_id,
			branch_id,
			recorded_by,
			category,
			description,
			amount,
			currency,
			payment_method,
			reference,
			expense_date,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12
		)
		RETURNING
			id,
			tenant_id,
			branch_id,
			recorded_by,
			category,
			description,
			amount::text,
			currency,
			payment_method,
			reference,
			expense_date,
			status,
			created_at,
			updated_at
	`

	expense := &Expense{}

	err := r.db.QueryRow(
		ctx,
		query,
		params.ID,
		params.TenantID,
		params.BranchID,
		params.RecordedBy,
		params.Category,
		params.Description,
		params.Amount,
		params.Currency,
		params.PaymentMethod,
		params.Reference,
		params.ExpenseDate,
		params.Status,
	).Scan(
		&expense.ID,
		&expense.TenantID,
		&expense.BranchID,
		&expense.RecordedBy,
		&expense.Category,
		&expense.Description,
		&expense.Amount,
		&expense.Currency,
		&expense.PaymentMethod,
		&expense.Reference,
		&expense.ExpenseDate,
		&expense.Status,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	id string,
) (*Expense, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			branch_id,
			recorded_by,
			category,
			description,
			amount::text,
			currency,
			payment_method,
			reference,
			expense_date,
			status,
			created_at,
			updated_at
		FROM expenses
		WHERE tenant_id = $1
		  AND id = $2
	`

	expense := &Expense{}

	err := r.db.QueryRow(
		ctx,
		query,
		tenantID,
		id,
	).Scan(
		&expense.ID,
		&expense.TenantID,
		&expense.BranchID,
		&expense.RecordedBy,
		&expense.Category,
		&expense.Description,
		&expense.Amount,
		&expense.Currency,
		&expense.PaymentMethod,
		&expense.Reference,
		&expense.ExpenseDate,
		&expense.Status,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (r *Repository) ListForTenant(
	ctx context.Context,
	tenantID string,
	branchID *string,
) ([]Expense, int64, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			branch_id,
			recorded_by,
			category,
			description,
			amount::text,
			currency,
			payment_method,
			reference,
			expense_date,
			status,
			created_at,
			updated_at
		FROM expenses
		WHERE tenant_id = $1
		  AND ($2::uuid IS NULL OR branch_id = $2)
		ORDER BY expense_date DESC, created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	expenses := make([]Expense, 0)

	for rows.Next() {
		var item Expense

		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.BranchID,
			&item.RecordedBy,
			&item.Category,
			&item.Description,
			&item.Amount,
			&item.Currency,
			&item.PaymentMethod,
			&item.Reference,
			&item.ExpenseDate,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		expenses = append(expenses, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64

	countQuery := `
		SELECT COUNT(*)
		FROM expenses
		WHERE tenant_id = $1
		  AND ($2::uuid IS NULL OR branch_id = $2)
	`

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		tenantID,
		branchID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

func (r *Repository) ListForBranch(
	ctx context.Context,
	tenantID string,
	branchID string,
) ([]Expense, int64, error) {
	return r.ListForTenant(
		ctx,
		tenantID,
		&branchID,
	)
}

func (r *Repository) Exists(
	ctx context.Context,
	tenantID string,
	id string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM expenses
			WHERE tenant_id = $1
			  AND id = $2
		)
		`,
		tenantID,
		id,
	).Scan(&exists)

	return exists, err
}

func parseUUID(value string) error {
	_, err := uuid.Parse(value)
	return err
}
