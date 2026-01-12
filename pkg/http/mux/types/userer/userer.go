package userer

import (
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

type Userer interface {
	GetUser() *motmedelHttpTypes.HttpContextUser
}

type UsererFunction func() *motmedelHttpTypes.HttpContextUser

func (f UsererFunction) GetUser() *motmedelHttpTypes.HttpContextUser {
	return f()
}

func New(f func() *motmedelHttpTypes.HttpContextUser) Userer {
	return UsererFunction(f)
}

// TODO: In future, add something for getting client metadata (IP enrichment; geo, tags e.g.)
