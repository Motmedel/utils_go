package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	"github.com/Motmedel/utils_go/pkg/database/sql/types/tx_caller"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	sqlErrors "github.com/Motmedel/utils_go/pkg/sql/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func WithTx[T any](
	ctx context.Context,
	database *sql.DB,
	txCaller tx_caller.TxCaller[T],
) (T, error) {
	var zero T

	if err := ctx.Err(); err != nil {
		return zero, fmt.Errorf("context err: %w", err)
	}

	if database == nil {
		return zero, motmedelErrors.NewWithTrace(sqlErrors.ErrNilSqlDatabase)
	}

	if utils.IsNil(txCaller) {
		return zero, motmedelErrors.NewWithTrace(errors.New("nil tx caller"))
	}

	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return zero, motmedelErrors.NewWithTrace(fmt.Errorf("begin transaction: %w", err))
	}
	if transaction == nil {
		return zero, motmedelErrors.NewWithTrace(sqlErrors.ErrNilTx)
	}

	out, err := txCaller.Call(ctx, transaction)
	if err != nil {
		var rollbackErr error
		if rollbackErr = transaction.Rollback(); rollbackErr != nil {
			slog.ErrorContext(
				motmedelContext.WithErrorContextValue(
					ctx,
					motmedelErrors.NewWithTrace(fmt.Errorf("tx rollback: %w", rollbackErr), transaction),
				),
				"An error occurred when rolling back a transaction.",
			)
		}
		return zero, err
	}

	if err := transaction.Commit(); err != nil {
		return zero, motmedelErrors.NewWithTrace(fmt.Errorf("tx commit: %w", err), transaction)
	}

	return out, nil
}
