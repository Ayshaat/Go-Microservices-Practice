package repository

import (
	"database/sql"
	stdErrors "errors"
	"stocks/internal/errors"
	"stocks/internal/models"
)

type PostgresStockRepo struct {
	db *sql.DB
}

func NewPostgresStockRepo(db *sql.DB) *PostgresStockRepo {
	return &PostgresStockRepo{db: db}
}

func (r *PostgresStockRepo) getSKUInfo(sku uint32) (string, string, error) {
	var name, itemType string

	err := r.db.QueryRow("SELECT name, type FROM sku_info WHERE sku = $1", sku).Scan(&name, &itemType)
	if stdErrors.Is(err, sql.ErrNoRows) {
		return "", "", errors.ErrInvalidSKU
	}

	if err != nil {
		return "", "", err
	}

	return name, itemType, nil
}

func (r *PostgresStockRepo) Add(item models.StockItem) error {
	name, itemType, err := r.getSKUInfo(item.SKU)
	if err != nil {
		return err
	}

	item.Name = name
	item.Type = itemType

	var exists bool

	err = r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM stock_items WHERE sku = $1)", item.SKU).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return errors.ErrItemExists
	}

	_, err = r.db.Exec(`
		INSERT INTO stock_items (user_id, sku, name, type, price, count, location)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, item.UserID, item.SKU, item.Name, item.Type, item.Price, item.Count, item.Location)

	return err
}

func (r *PostgresStockRepo) Delete(sku uint32) error {
	res, err := r.db.Exec("DELETE FROM stock_items WHERE sku = $1", sku)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.ErrItemNotFound
	}

	return nil
}

func (r *PostgresStockRepo) GetBySKU(sku uint32) (models.StockItem, error) {
	var item models.StockItem
	err := r.db.QueryRow(`
		SELECT s.user_id, s.sku, i.name, i.type, s.price, s.count, s.location 
		FROM stock_items s
		JOIN sku_info i ON s.sku = i.sku
		WHERE s.sku = $1
	`, sku).Scan(&item.UserID, &item.SKU, &item.Name, &item.Type, &item.Price, &item.Count, &item.Location)

	if stdErrors.Is(err, sql.ErrNoRows) {
		return models.StockItem{}, errors.ErrItemNotFound
	}

	return item, err
}

func (r *PostgresStockRepo) GetSKUInfo(sku uint32) (string, string, error) {
	var name, typ string
	err := r.db.QueryRow(
		`SELECT name, type FROM sku_info WHERE sku = $1`, sku,
	).Scan(&name, &typ)

	if stdErrors.Is(err, sql.ErrNoRows) {
		return "", "", errors.ErrItemNotFound
	}

	return name, typ, err
}

func (r *PostgresStockRepo) ListByLocation(location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	offset := (currentPage - 1) * pageSize
	rows, err := r.db.Query(`
		SELECT s.user_id, s.sku, i.name, i.type, s.price, s.count, s.location 
		FROM stock_items s
		JOIN sku_info i ON s.sku = i.sku
		WHERE s.location = $1
		ORDER BY s.sku 
		LIMIT $2 OFFSET $3
	`, location, pageSize, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]models.StockItem, 0)

	for rows.Next() {
		var row StockItemRow

		err := rows.Scan(&row.UserID, &row.SKU, &row.Name, &row.Type, &row.Price, &row.Count, &row.Location)
		if err != nil {
			return nil, err
		}

		items = append(items, row.ToDomain())
	}

	return items, rows.Err()
}
