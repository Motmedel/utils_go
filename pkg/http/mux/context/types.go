package context

type requestIdContextType struct{}
type httpContextContextType struct{}

var RequestIdContextKey requestIdContextType
var HttpContextContextKey httpContextContextType
