package interfaces

import (
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

type Userer interface {
	GetUser() *motmedelHttpTypes.HttpContextUser
}

// TODO: In future, add something for getting client metadata (IP enrichment; geo, tags e.g.)
