package errors

import "errors"

var (
	ErrNilSqlDatabase = errors.New("nil sql database")
	ErrNilRows        = errors.New("nil rows")
	ErrNilRow         = errors.New("nil row")
	ErrNilTx          = errors.New("nil tx")

	ErrEmptyQuery            = errors.New("empty query")
	ErrNilTxCaller           = errors.New("nil tx caller")
	ErrNilTxAuthorizer       = errors.New("nil tx authorizer")
	ErrNilAuthorizedTxCaller = errors.New("nil authorized tx caller")
)
