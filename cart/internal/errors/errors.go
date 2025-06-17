package errors

import "errors"

var (
	ErrInvalidSKU       = errors.New("invalid SKU — not registered")
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrCartItemExists   = errors.New("cart item already exists")
)
