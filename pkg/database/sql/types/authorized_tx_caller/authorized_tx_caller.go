package authorized_tx_caller

import (
	"context"
	"database/sql"
	"fmt"

	sqlErrors "github.com/Motmedel/utils_go/pkg/database/sql/errors"
	"github.com/Motmedel/utils_go/pkg/database/sql/types/tx_authorizer"
	"github.com/Motmedel/utils_go/pkg/database/sql/types/tx_caller"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type AuthorizedTxCaller[T any] struct {
	Id string
	tx_caller.TxCaller[T]
	tx_authorizer.TxAuthorizer
}

func (c *AuthorizedTxCaller[T]) Call(ctx context.Context, tx *sql.Tx) (T, error) {
	var zero T

	if utils.IsNil(c.TxCaller) {
		return zero, motmedelErrors.NewWithTrace(sqlErrors.ErrNilTxCaller)
	}

	if utils.IsNil(c.TxAuthorizer) {
		return zero, motmedelErrors.NewWithTrace(sqlErrors.ErrNilTxCaller)
	}

	authorized, err := c.TxAuthorizer.Authorized(ctx, c.Id, tx)
	if err != nil {
		return zero, fmt.Errorf("tx authorizer authorized: %w", err)
	}
	if !authorized {
		return zero, motmedelErrors.ErrUnauthorized
	}

	out, err := c.TxCaller.Call(ctx, tx)
	if err != nil {
		return zero, fmt.Errorf("tx caller: %w", err)
	}

	return out, nil
}

func New[T any](id string, txCaller tx_caller.TxCaller[T], txAuthorizer tx_authorizer.TxAuthorizer) *AuthorizedTxCaller[T] {
	return &AuthorizedTxCaller[T]{Id: id, TxCaller: txCaller, TxAuthorizer: txAuthorizer}
}
