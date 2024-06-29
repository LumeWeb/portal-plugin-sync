package api

import (
	_ "embed"
	"encoding/hex"
	"github.com/gorilla/mux"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/portal-plugin-sync/types"
	"go.lumeweb.com/portal/config"
	"go.lumeweb.com/portal/core"
	"go.lumeweb.com/portal/middleware"
	"go.lumeweb.com/portal/middleware/swagger"
	"net/http"
)

const subdomain = "sync"

//go:embed swagger.yaml
var swagSpec []byte

var _ core.API = (*SyncAPI)(nil)

type SyncAPI struct {
	ctx    core.Context
	config config.Manager
	logger *core.Logger
	sync   types.SyncService
	user   core.UserService
}

func NewSyncAPI() (*SyncAPI, []core.ContextBuilderOption, error) {

	api := &SyncAPI{}

	opts := core.ContextOptions(
		core.ContextWithStartupFunc(func(ctx core.Context) error {
			api.ctx = ctx
			api.config = ctx.Config()
			api.logger = ctx.Logger()
			api.sync = ctx.Service(types.SYNC_SERVICE).(types.SyncService)
			api.user = ctx.Service(core.USER_SERVICE).(core.UserService)
			return nil
		}),
	)

	return api, opts, nil
}

func (s *SyncAPI) Config() config.APIConfig {
	return nil
}

func (s *SyncAPI) Name() string {
	return "sync"
}

func (s *SyncAPI) AuthTokenName() string {
	return core.AUTH_COOKIE_NAME
}

func (s *SyncAPI) Routes() (*mux.Router, error) {
	authMiddlewareOpts := middleware.AuthMiddlewareOptions{
		Context: s.ctx,
		Purpose: core.JWTPurposeLogin,
	}

	authMw := middleware.AuthMiddleware(authMiddlewareOpts)

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

	err := swagger.Swagger(swagSpec, router)
	if err != nil {
		return err
	}

	authMiddlewareOpts := middleware.AuthMiddlewareOptions{
		Context: s.ctx,
	}

	authMw := middleware.AuthMiddleware(authMiddlewareOpts)

	router.HandleFunc("/api/log/key", s.logKey).Methods("GET")
	router.HandleFunc("/api/import", s.objectImport).Methods("POST").Use(authMw)

	return nil
}

func (s *SyncAPI) Subdomain() string {
	return subdomain
}
