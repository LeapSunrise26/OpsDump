package cmd

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"ops-dump/internal/controller"
)

// Register binds all module controllers to the given HTTP server using the
// strict-router (object registration) style. Route paths/methods are defined
// via g.Meta in the api/v1 request structs.
func Register(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(controller.MiddlewareHandlerResponse)

		// Public system endpoints (no auth).
		group.Bind(
			controller.NewSystem(),
		)

		// Server-rendered pages (require login).
		group.Group("/", func(g *ghttp.RouterGroup) {
			g.Middleware(controller.Auth)
			page := controller.NewPage()
			g.GET("/", page.Dashboard)
			g.GET("/login", page.LoginPage)
			g.GET("/dashboard", page.Dashboard)
			g.GET("/jobs", page.Jobs)
			g.GET("/runs", page.Runs)
		})

		// JSON API. Login is allowed through; other endpoints require auth.
		group.Group("/api", func(g *ghttp.RouterGroup) {
			g.Middleware(controller.Auth)
			g.Bind(
				controller.NewAuth(),
				controller.NewJob(),
				controller.NewNotify(),
				controller.NewDashboard(),
			)
		})
	})
}
