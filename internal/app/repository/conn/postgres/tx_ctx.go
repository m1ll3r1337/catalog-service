package rcpostgres

import (
	"context"

	"github.com/uptrace/bun"
)

type contextKeyTx struct{}

// getTxFromContext извлекает bun.Tx из контекста.
// Если транзакции нет, возвращает zero-value bun.Tx.
func getTxFromContext(ctx context.Context) bun.Tx {
	if v := ctx.Value(contextKeyTx{}); v != nil {
		if tx, ok := v.(bun.Tx); ok {
			return tx
		}
	}
	return bun.Tx{}
}

// setTxToContext сохраняет bun.Tx в контекст.
func setTxToContext(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, contextKeyTx{}, tx)
}
