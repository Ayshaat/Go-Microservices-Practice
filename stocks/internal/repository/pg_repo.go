package repository

import (
	"database/sql"
	"stocks/internal/models"
)

type pgStockRepository struct {
	db *sql.DB
}

func NewPGStockRepository(db *sql.DB) *pgStockRepository {
	return &pgStockRepository{db: db}
}

func (r *pgStockRepository) Add(item models.StockItem) error {
	var exists bool

	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM stock_items WHERE sku = $1)", item.SKU).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return ErrItemExists
	}

	_, err = r.db.Exec(
		`INSERT INTO stock_items (sku, name, type, price, count, location) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		item.SKU, item.Name, item.Type, item.Price, item.Count, item.Location)

	return err
}

func (r *pgStockRepository) Delete(sku uint32) error {
	res, err := r.db.Exec("DELETE FROM stock_items WHERE sku = $1", sku)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *pgStockRepository) GetBySKU(sku uint32) (models.StockItem, error) {
	var item models.StockItem
	err := r.db.QueryRow(
		`SELECT sku, name, type, price, count, location 
		 FROM stock_items WHERE sku = $1`, sku).
		Scan(&item.SKU, &item.Name, &item.Type, &item.Price, &item.Count, &item.Location)

	if err == sql.ErrNoRows {
		return models.StockItem{}, ErrItemNotFound
	}

	return item, err
}

func (r *pgStockRepository) ListByLocation(location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	offset := (currentPage - 1) * pageSize

	rows, err := r.db.Query(
		`SELECT sku, name, type, price, count, location 
		 FROM stock_items WHERE location = $1 
		 ORDER BY sku 
		 LIMIT $2 OFFSET $3`, location, pageSize, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []models.StockItem

	for rows.Next() {
		var item models.StockItem

		err := rows.Scan(&item.SKU, &item.Name, &item.Type, &item.Price, &item.Count, &item.Location)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
