package define

import (
	"github.com/go-co-op/gocron/v2"
	"go.lumeweb.com/portal-plugin-sync/internal/metadata"
)

const CronTaskVerifyObjectName = "SyncVerifyObject"
const CronTaskUploadObjectName = "SyncUploadObject"
const CronTaskScanObjectsName = "SyncScanObjects"

type CronTaskVerifyObjectArgs struct {
	Hash       []byte              `json:"hash"`
	Object     []metadata.FileMeta `json:"object"`
	UploaderID uint64              `json:"uploader_id"`
}

type CronTaskUploadObjectArgs struct {
	Hash       []byte `json:"hash"`
	Protocol   string `json:"protocol"`
	Size       uint64 `json:"size"`
	UploaderID uint64 `json:"uploader_id"`
}

func CronTaskUploadObjectArgsFactory() any {
	return &CronTaskUploadObjectArgs{}
}

func CronTaskScanObjectsDefinition() gocron.JobDefinition {
	return gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0)))
}
