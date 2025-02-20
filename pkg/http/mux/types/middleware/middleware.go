package middleware

import "net/http"

type Middleware func(request *http.Request) *http.Request
