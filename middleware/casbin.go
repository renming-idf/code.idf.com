package middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/spf13/cast"
	"net/http"
	"strconv"

	"github.com/casbin/casbin/v2"
)

/*
	Updated for the casbin 2.x released 3 days ago.
	2019-07-15
*/

// New returns the auth service which receives a casbin enforcer.
//
// Adapt with its `Wrapper` for the entire application
// or its `ServeHTTP` for specific routes or parties.
func New(e *casbin.Enforcer) *Casbin {
	return &Casbin{enforcer: e}
}

// ServeHTTP is the iris compatible casbin handler which should be passed to specific routes or parties.
// Usage:
// [...]
// app.Get("/dataset1/resource1", casbinMiddleware.ServeHTTP, myHandler)
// [...]
func (c *Casbin) ServeHTTP(ctx iris.Context) {
	m, ok := ParseToken(ctx.GetHeader("token"))
	if !ok {
		ctx.StatusCode(http.StatusForbidden) // Status Forbidden
		ctx.StopExecution()
		return
	}
	aid := cast.ToUint(m["aid"])
	if aid < 1 {
		ctx.StatusCode(http.StatusForbidden) // Status Forbidden
		ctx.StopExecution()
		return
	}
	if !c.Check(ctx.Request(), strconv.FormatUint(uint64(aid), 10)) {
		ctx.StatusCode(http.StatusForbidden) // Status Forbidden
		ctx.StopExecution()
		return
	} else {
		ctx.Values().Set("auth_user_id", aid)
	}

	ctx.Next()
}

// Casbin is the auth services which contains the casbin enforcer.
type Casbin struct {
	enforcer *casbin.Enforcer
}

// Check checks the username, request's method and path and
// returns true if permission grandted otherwise false.
func (c *Casbin) Check(r *http.Request, userId string) bool {
	method := r.Method
	path := r.URL.Path
	ok, _ := c.enforcer.Enforce(userId, path, method)
	return ok
}
