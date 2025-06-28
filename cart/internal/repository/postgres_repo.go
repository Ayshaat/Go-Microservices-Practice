package repository

import (
	"cart/internal/errors"
	"cart/internal/models"
	"context"
	"database/sql"
	stdErrors "errors"
)

type PostgresCartRepo struct {
	db *sql.DB
}

func NewPostgresCartRepo(db *sql.DB) *PostgresCartRepo {
	return &PostgresCartRepo{db: db}
}

func (r *PostgresCartRepo) GetSKUInfo(ctx context.Context, sku uint32) (string, string, error) {
	var name, typ string

	err := r.db.QueryRowContext(ctx, "SELECT name, type FROM sku_info WHERE sku = $1", sku).Scan(&name, &typ)
	if stdErrors.Is(err, sql.ErrNoRows) {
		return "", "", errors.ErrInvalidSKU
	}

	if err != nil {
		return "", "", err
	}

	return name, typ, nil
}

func (r *PostgresCartRepo) Add(ctx context.Context, item models.CartItem) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO cart_items (user_id, sku, count) VALUES ($1, $2, $3)`,
		item.UserID, item.SKU, item.Count)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresCartRepo) Upsert(ctx context.Context, item models.CartItem) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, sku, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, sku)
		DO UPDATE SET 
			count = cart_items.count + EXCLUDED.count
	`, item.UserID, item.SKU, item.Count)

	return err
}

func (r *PostgresCartRepo) Delete(ctx context.Context, userID int64, sku uint32) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1 AND sku = $2`, userID, sku)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.ErrCartItemNotFound
	}

	return nil
}

func (r *PostgresCartRepo) List(ctx context.Context, userID int64) ([]models.CartItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, sku, count FROM cart_items WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CartItem

	for rows.Next() {
		var row CartItemRow

		if err = rows.Scan(&row.UserID, &row.SKU, &row.Count); err != nil {
			return nil, err
		}

		items = append(items, row.ToDomain())
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresCartRepo) Clear(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}
