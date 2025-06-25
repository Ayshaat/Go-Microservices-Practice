package repository

import (
	"context"
	"database/sql"
	stdErrors "errors"
	"stocks/internal/errors"
	"stocks/internal/models"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
)

type PostgresStockRepo struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewPostgresStockRepo(db *sqlx.DB, getter *trmsqlx.CtxGetter) *PostgresStockRepo {
	return &PostgresStockRepo{db: db, getter: getter}
}

func (r *PostgresStockRepo) itemExists(ctx context.Context, sku uint32) (bool, error) {
	var exists bool
	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM stock_items WHERE sku = $1)", sku).Scan(&exists)

	return exists, err
}

func (r *PostgresStockRepo) insertStockItem(ctx context.Context, item models.StockItem) error {
	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, `
		INSERT INTO stock_items (user_id, sku, price, count, location)
		VALUES ($1, $2, $3, $4, $5)
	`, item.UserID, item.SKU, item.Price, item.Count, item.Location)

	return err
}

func (r *PostgresStockRepo) Add(ctx context.Context, item models.StockItem) error {
	exists, err := r.itemExists(ctx, item.SKU)
	if err != nil {
		return err
	}

	if exists {
		return errors.ErrItemExists
	}

	return r.insertStockItem(ctx, item)
}

func (r *PostgresStockRepo) Delete(ctx context.Context, sku uint32) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM stock_items WHERE sku = $1", sku)
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

func (r *PostgresStockRepo) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	var item models.StockItem
	err := r.db.QueryRowContext(ctx, `
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

func (r *PostgresStockRepo) GetSKUInfo(ctx context.Context, sku uint32) (string, string, error) {
	var name, typ string
	err := r.db.QueryRowContext(ctx,
		`SELECT name, type FROM sku_info WHERE sku = $1`,
		sku,
	).Scan(&name, &typ)

	if stdErrors.Is(err, sql.ErrNoRows) {
		return "", "", errors.ErrItemNotFound
	}

	return name, typ, err
}

func (r *PostgresStockRepo) ListByLocation(ctx context.Context, location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	offset := (currentPage - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
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
