package errors

import "errors"

var (
	ErrNilSqlDatabase = errors.New("nil sql database")
	ErrNilRows        = errors.New("nil rows")
	ErrNilRow         = errors.New("nil row")
	ErrNilTx          = errors.New("nil tx")
)
