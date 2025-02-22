package context

type requestIdContextType struct{}

var RequestIdContextKey = &requestIdContextType{}

type httpContextContextType struct{}

var HttpContextContextKey httpContextContextType
