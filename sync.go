package sync

import (
	"go.lumeweb.com/portal-plugin-sync/internal/api"
	"go.lumeweb.com/portal-plugin-sync/internal/service"
	"go.lumeweb.com/portal-plugin-sync/types"
	"go.lumeweb.com/portal/core"
)

func init() {
	core.RegisterPlugin(core.PluginInfo{
		ID: "sync",
		API: func() (core.API, []core.ContextBuilderOption, error) {
			return api.NewSyncAPI()
		},
		Services: func() ([]core.ServiceInfo, error) {
			return []core.ServiceInfo{
				{
					ID: types.SYNC_SERVICE,
					Factory: func() (core.Service, []core.ContextBuilderOption, error) {
						return service.NewSyncService()
					},
					Depends: []string{core.RENTER_SERVICE, core.STORAGE_SERVICE, core.METADATA_SERVICE, core.CRON_SERVICE},
				},
			}, nil
		},
	})
}
