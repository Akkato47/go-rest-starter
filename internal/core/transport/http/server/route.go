package server

import "github.com/gin-gonic/gin"

type Route struct {
	Method     string
	Path       string
	Handler    gin.HandlerFunc
	Middleware []gin.HandlerFunc
}

type RouterGroup struct {
	prefix     string
	middleware []gin.HandlerFunc
	routes     []Route
}

func NewRouterGroup(prefix string, middleware ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{prefix: prefix, middleware: middleware}
}

func (g *RouterGroup) AddRoutes(routes ...Route) {
	g.routes = append(g.routes, routes...)
}

func (g *RouterGroup) register(engine *gin.Engine) {
	group := engine.Group(g.prefix, g.middleware...)
	for _, route := range g.routes {
		handlers := make([]gin.HandlerFunc, 0, len(route.Middleware)+1)
		handlers = append(handlers, route.Middleware...)
		handlers = append(handlers, route.Handler)
		group.Handle(route.Method, route.Path, handlers...)
	}
}
