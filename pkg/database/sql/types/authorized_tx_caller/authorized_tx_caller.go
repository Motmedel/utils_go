package authorized_tx_caller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Motmedel/utils_go/pkg/database/sql/types/tx_authorizer"
	"github.com/Motmedel/utils_go/pkg/database/sql/types/tx_caller"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type AuthorizedTxCaller[T any] struct {
	tx_caller.TxCaller[T]
	tx_authorizer.TxAuthorizer
}

func (c *AuthorizedTxCaller[T]) Call(ctx context.Context, tx *sql.Tx) (T, error) {
	var zero T

	if utils.IsNil(c.TxCaller) {
		// TODO: Fix error
		return zero, motmedelErrors.NewWithTrace(errors.New("nil tx caller"))
	}

	if utils.IsNil(c.TxAuthorizer) {
		// TODO: Fix error
		return zero, motmedelErrors.NewWithTrace(errors.New("nil tx caller"))
	}

	authorized, err := c.TxAuthorizer.Authorized(ctx, tx)
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

func New[T any](txCaller tx_caller.TxCaller[T], txAuthorizer tx_authorizer.TxAuthorizer) *AuthorizedTxCaller[T] {
	return &AuthorizedTxCaller[T]{TxCaller: txCaller, TxAuthorizer: txAuthorizer}
}
