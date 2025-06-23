package repository

import (
	"cart/internal/errors"
	"cart/internal/models"
	"database/sql"
)

type PostgresCartRepo struct {
	db *sql.DB
}

func NewPostgresCartRepo(db *sql.DB) *PostgresCartRepo {
	return &PostgresCartRepo{db: db}
}

func (r *PostgresCartRepo) GetSKUInfo(sku uint32) (string, string, error) {
	var name, typ string
	err := r.db.QueryRow("SELECT name, type FROM sku_info WHERE sku = $1", sku).Scan(&name, &typ)
	if err == sql.ErrNoRows {
		return "", "", errors.ErrInvalidSKU
	}
	if err != nil {
		return "", "", err
	}
	return name, typ, nil
}

func (r *PostgresCartRepo) Add(item models.CartItem) error {
	_, err := r.db.Exec(`INSERT INTO cart_items (user_id, sku, count) VALUES ($1, $2, $3)`,
		item.UserID, item.SKU, item.Count)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresCartRepo) Delete(userID int64, sku uint32) error {
	res, err := r.db.Exec(`DELETE FROM cart_items WHERE user_id = $1 AND sku = $2`, userID, sku)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.ErrCartItemNotFound
	}

	return nil
}

func (r *PostgresCartRepo) List(userID int64) ([]models.CartItem, error) {
	rows, err := r.db.Query(`SELECT user_id, sku, count FROM cart_items WHERE user_id = $1`, userID)
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

func (r *PostgresCartRepo) Clear(userID int64) error {
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}
