package sync

import (
	"github.com/LumeWeb/portal/core"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/LumeWeb/portal/middleware"
)

const (
	authCookieName = core.AUTH_COOKIE_NAME
	authQueryParam = "auth_token"
)

func findToken(r *http.Request) string {
	return middleware.FindAuthToken(r, authCookieName, authQueryParam)
}

func authMiddleware(options middleware.AuthMiddlewareOptions) mux.MiddlewareFunc {
	options.FindToken = findToken
	return middleware.AuthMiddleware(options)
}
