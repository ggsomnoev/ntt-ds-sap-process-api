package txctx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type txKeyType struct{}

var txKey = txKeyType{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func GetTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("no transaction provided in context")
	}
	return tx, nil
}