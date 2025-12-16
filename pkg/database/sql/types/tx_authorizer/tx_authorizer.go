package tx_authorizer

import (
	"context"
	"database/sql"
)

type TxAuthorizer interface {
	Authorized(context.Context, string, *sql.Tx) (bool, error)
}

type TxAuthorizerFunction func(context.Context, string, *sql.Tx) (bool, error)

func (f TxAuthorizerFunction) Authorized(ctx context.Context, id string, tx *sql.Tx) (bool, error) {
	return f(ctx, id, tx)
}

func New(fn func(context.Context, string, *sql.Tx) (bool, error)) TxAuthorizer {
	return TxAuthorizerFunction(fn)
}
