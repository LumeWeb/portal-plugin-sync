package cron

import (
	"go.lumeweb.com/portal-plugin-sync/internal/cron/define"
	"go.lumeweb.com/portal-plugin-sync/internal/cron/tasks"
	"go.lumeweb.com/portal/core"
)

var _ core.Cronable = (*Cron)(nil)

type Cron struct {
	ctx core.Context
}

func (c Cron) RegisterTasks(crn core.CronService) error {
	crn.RegisterTask(define.CronTaskVerifyObjectName, tasks.CronTaskVerifyObject, core.CronTaskDefinitionOneTimeJob, core.CronTaskNoArgsFactory)
	crn.RegisterTask(define.CronTaskUploadObjectName, tasks.CronTaskUploadObject, core.CronTaskDefinitionOneTimeJob, define.CronTaskUploadObjectArgsFactory)
	crn.RegisterTask(define.CronTaskScanObjectsName, tasks.CronTaskScanObjects, define.CronTaskScanObjectsDefinition, core.CronTaskNoArgsFactory)
	return nil
}

func (c Cron) ScheduleJobs(_ core.CronService) error {
	return nil
}

func NewCron(ctx core.Context) *Cron {
	return &Cron{ctx: ctx}
}
