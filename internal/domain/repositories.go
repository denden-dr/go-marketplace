package domain

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Pool interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}
