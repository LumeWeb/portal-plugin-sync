package types

import "go.lumeweb.com/portal/core"

const SYNC_SERVICE = "sync"

type SyncProtocol interface {
	Name() string
	EncodeFileName([]byte) string
	ValidIdentifier(string) bool
	HashFromIdentifier(string) ([]byte, error)
	StorageProtocol() core.StorageProtocol
}
type SyncService interface {
	Update(upload core.UploadMetadata) error
	LogKey() []byte
	Import(object string, uploaderID uint64) error
	Enabled() bool

	core.Service
}
