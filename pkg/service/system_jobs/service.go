package systemjobs

import (
	analyticsJob "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/analytics_job"
	nodeJob "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/node_job"
	sunriseJob "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/sunrise_job"
)

// ReloadSystemJobs func
func ReloadSystemJobs() {
	sunriseJob.ReloadJob()
	analyticsJob.ReloadJob()
	nodeJob.ReloadNodeStateVerifyJob()
}
