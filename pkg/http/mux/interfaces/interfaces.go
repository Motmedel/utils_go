package interfaces

import "github.com/Motmedel/ecs_go/ecs"

type Userer interface {
	GetUser() *ecs.User
}

// TODO: In future, add something for getting client metadata (IP enrichment; geo, tags e.g.)
