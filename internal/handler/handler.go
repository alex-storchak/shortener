package handler

import "net/http"

type Route struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	Middlewares []func(http.Handler) http.Handler
}

type Handler interface {
	Routes() []Route
}

type Handlers []Handler

type Middlewares []func(http.Handler) http.Handler
