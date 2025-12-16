package tx_authorizer

import (
	"context"
	"database/sql"
)

type TxAuthorizer interface {
	Authorized(context.Context, *sql.Tx) (bool, error)
}

type TxAuthorizerFunction func(context.Context, *sql.Tx) (bool, error)

func (f TxAuthorizerFunction) Authorized(ctx context.Context, tx *sql.Tx) (bool, error) {
	return f(ctx, tx)
}

func New(fn func(context.Context, *sql.Tx) (bool, error)) TxAuthorizer {
	return TxAuthorizerFunction(fn)
}
