package extrav2

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"go.uber.org/zap"
)

// AdminV2 struct
type AdminV2 struct {
	api    api.BucketsAPI
	orgID  string
	bucket string
	ctx    context.Context
}

// NewAdminClient returns influxdb admin client
func NewAdminClient(ctx context.Context, api api.BucketsAPI, orgID, bucket string) *AdminV2 {
	return &AdminV2{ctx: ctx, api: api, orgID: orgID, bucket: bucket}
}

// IsBucketAvailable returns the availability of the database
func (av2 *AdminV2) IsBucketAvailable() (bool, error) {
	bucketDomain, err := av2.api.FindBucketByID(av2.ctx, av2.bucket)
	if err != nil {
		return false, err
	}
	return *bucketDomain.Id == av2.bucket, nil
}

// CreateBucket adds a bucket to influxdb
func (av2 *AdminV2) CreateBucket() error {
	status, err := av2.IsBucketAvailable()
	if err != nil {
		return err
	}
	if status {
		return nil
	}

	bucketDomain := &domain.Bucket{Id: &av2.bucket, OrgID: &av2.orgID}
	_, err = av2.api.CreateBucket(av2.ctx, bucketDomain)
	if err != nil {
		return err
	}

	zap.L().Info("metrics bucket created", zap.String("organization", av2.orgID), zap.String("bucket", av2.bucket))
	return nil
}
