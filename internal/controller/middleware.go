package controller

import (
	"mime"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"ops-dump/internal/consts"
)

// streamContentType lists content types that must never be wrapped into the
// unified JSON response. Mirrors the unexported list used by
// ghttp.MiddlewareHandlerResponse.
var streamContentType = []string{
	"text/event-stream",
	"application/octet-stream",
	"multipart/x-mixed-replace",
}

// Auth guards admin pages and /api endpoints. Whitelisted paths (login page,
// dashboard root and the login API) pass through; unauthenticated API calls
// receive an error that the response middleware turns into {code:1,...};
// unauthenticated page calls are redirected to the login page.
func Auth(r *ghttp.Request) {
	if r.Session.MustGet(consts.SessionUserId).Int64() > 0 {
		r.Middleware.Next()
		return
	}
	path := r.URL.Path
	if path == consts.PageLogin || path == "/" {
		r.Middleware.Next()
		return
	}
	if len(path) >= 4 && path[:4] == "/api" {
		if path == "/api/login" {
			r.Middleware.Next()
			return
		}
		r.SetError(gerror.New("未登录"))
		return
	}
	r.Response.RedirectTo(consts.PageLogin)
}

// MiddlewareHandlerResponse is a drop-in replacement for
// ghttp.MiddlewareHandlerResponse that additionally treats 3xx (redirect)
// responses as a normal outcome instead of an error.
//
// Why: GoFrame's built-in middleware classifies any non-200/404/403 status as
// gcode.CodeUnknown when the handler chain wrote no buffer. Auth's
// RedirectTo("/login") produces exactly that case (302 + Location header, no
// body), so every unauthenticated page hit was logged as a misleading
// "Unknown Error" in the server error log even though the redirect worked.
// With this middleware, 3xx responses are returned as-is and never recorded
// as errors.
func MiddlewareHandlerResponse(r *ghttp.Request) {
	r.Middleware.Next()

	// There's custom buffer content, it then exits current handler.
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	// Redirect responses (e.g. Auth -> /login) are normal control flow.
	if r.Response.Status >= 300 && r.Response.Status < 400 {
		return
	}

	// It does not output common response content if it is stream response.
	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	for _, ct := range streamContentType {
		if mediaType == ct {
			return
		}
	}

	var (
		msg  string
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		msg = err.Error()
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden:
				code = gcode.CodeNotAuthorized
			default:
				code = gcode.CodeUnknown
			}
			// It creates an error as it can be retrieved by other middlewares.
			err = gerror.NewCode(code, msg)
			r.SetError(err)
		} else {
			code = gcode.CodeOK
		}
		msg = code.Message()
	}

	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    code.Code(),
		Message: msg,
		Data:    res,
	})
}
