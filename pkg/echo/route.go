package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type RouteInitializer interface {
	SetupRoute()
}

type restRouter struct {
	e      *echo.Group
	prefix *string
}

type RouteInfo struct {
	ServiceID string
	Remarks   string
}

var ServiceIDRoute = map[string]*RouteInfo{}

func NewRestRoute(e *echo.Echo) RestRoute {
	return &restRouter{e: e.Group("")}
}

func (svc *restRouter) Group(prefix string) RestRoute {
	if svc.prefix == nil {
		svc.prefix = new(string)
	}
	newPrefix := *svc.prefix + prefix
	return &restRouter{e: svc.e.Group(prefix), prefix: &newPrefix}
}

func (svc *restRouter) Use(m ...echo.MiddlewareFunc) {
	svc.e.Use(m...)
}

func (svc *restRouter) Post(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	svc.register(http.MethodPost, path, serviceID, remarks)
	return svc.e.POST(path, handler, m...)
}

func (svc *restRouter) Get(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	svc.register(http.MethodGet, path, serviceID, remarks)
	return svc.e.GET(path, handler, m...)
}

func (svc *restRouter) Put(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	svc.register(http.MethodPut, path, serviceID, remarks)
	return svc.e.PUT(path, handler, m...)
}

func (svc *restRouter) Delete(serviceID, remarks, path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	svc.register(http.MethodDelete, path, serviceID, remarks)
	return svc.e.DELETE(path, handler, m...)
}

func (svc *restRouter) register(method, path, serviceID, remarks string) {
	if svc.prefix == nil {
		svc.prefix = new(string)
	}
	ServiceIDRoute[method+*svc.prefix+path] = &RouteInfo{ServiceID: serviceID, Remarks: remarks}
}
