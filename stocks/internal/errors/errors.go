package errors

import "errors"

var (
	ErrItemExists   = errors.New("item already exists")
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidSKU   = errors.New("invalid SKU — not registered")
)
