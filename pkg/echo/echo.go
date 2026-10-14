package echo

import "github.com/labstack/echo/v4"

type (
	Context        = echo.Context
	HandlerFunc    = echo.HandlerFunc
	MiddlewareFunc = echo.MiddlewareFunc
	Route          = echo.Route
	Echo           = echo.Echo
	Group          = echo.Group
)

type RestRoute interface {
	Group(prefix string) RestRoute
	Post(serviceID, remarks, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route
	Get(serviceID, remarks, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route
	Put(serviceID, remarks, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route
	Delete(serviceID, remarks, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route
	Use(m ...MiddlewareFunc)
}

