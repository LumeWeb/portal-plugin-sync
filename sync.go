package sync

import (
	"encoding/hex"
	"github.com/LumeWeb/httputil"
	"github.com/LumeWeb/portal/config"
	"github.com/LumeWeb/portal/core"
	"github.com/LumeWeb/portal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

const subdomain = "sync"

var _ core.API = (*SyncAPI)(nil)

func init() {
	core.RegisterPlugin(factory)
}

type SyncAPI struct {
	ctx    core.Context
	config config.Manager
	logger core.Logger
	sync   core.SyncService
	user   core.UserService
}

func NewSync(ctx core.Context) *SyncAPI {
	return &SyncAPI{
		ctx:    ctx,
		config: ctx.Config(),
		sync:   ctx.Services().Sync(),
		user:   ctx.Services().User(),
	}
}

func (s *SyncAPI) Name() string {
	return "sync"
}

func (s *SyncAPI) AuthTokenName() string {
	return authCookieName
}

func (s *SyncAPI) Routes() (*mux.Router, error) {
	authMiddlewareOpts := middleware.AuthMiddlewareOptions{
		Context: s.ctx,
		Purpose: core.JWTPurposeLogin,
	}

	authMw := authMiddleware(authMiddlewareOpts)

	router := mux.NewRouter()

	router.HandleFunc("/api/log/key", s.logKey).Methods("GET")
	router.HandleFunc("/api/import", s.objectImport).Methods("POST").Use(authMw)

	return router, nil
}

func (s *SyncAPI) logKey(w http.ResponseWriter, r *http.Request) {
	ctx := httputil.Context(r, w)
	keyHex := hex.EncodeToString(s.sync.LogKey())

	response := LogKeyResponse{
		Key: keyHex,
	}

	ctx.Encode(response)
}

func (s *SyncAPI) objectImport(w http.ResponseWriter, r *http.Request) {
	ctx := httputil.Context(r, w)

	var req ObjectImportRequest
	err := ctx.Decode(&req)
	if err != nil {
		return
	}

	user := middleware.GetUserFromContext(r.Context())

	err = s.sync.Import(req.Object, uint64(user))
	if err != nil {
		_ = ctx.Error(err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *SyncAPI) Configure(router *mux.Router) error {
	authMiddlewareOpts := middleware.AuthMiddlewareOptions{
		Context: s.ctx,
	}

	authMw := authMiddleware(authMiddlewareOpts)

	router.HandleFunc("/api/log/key", s.logKey).Methods("GET")
	router.HandleFunc("/api/import", s.objectImport).Methods("POST").Use(authMw)

	return nil
}

func (s *SyncAPI) Subdomain() string {
	return subdomain
}

func factory() core.PluginInfo {
	return core.PluginInfo{
		ID: "sync",
		GetAPI: func(ctx *core.Context) (core.API, error) {
			if ctx.Services().Sync().Enabled() {
				return NewSync(*ctx), nil
			}

			return nil, nil
		},
	}
}
