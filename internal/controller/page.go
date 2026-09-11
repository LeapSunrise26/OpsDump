package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// PageController renders server-side HTML pages.
type PageController struct{}

func NewPage() *PageController { return &PageController{} }

func (c *PageController) LoginPage(r *ghttp.Request) {
	r.Response.WriteTpl("login.html")
}

func (c *PageController) Dashboard(r *ghttp.Request) {
	r.Response.WriteTpl("dashboard.html")
}

func (c *PageController) Jobs(r *ghttp.Request) {
	r.Response.WriteTpl("jobs.html")
}

func (c *PageController) Runs(r *ghttp.Request) {
	r.Response.WriteTpl("runs.html")
}
