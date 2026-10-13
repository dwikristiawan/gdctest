package echo

import "github.com/labstack/echo/v4"

type RestRoute interface {
	Group(prefix string) RestRoute
	Post(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Get(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Put(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Delete(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(m ...echo.MiddlewareFunc)
}
