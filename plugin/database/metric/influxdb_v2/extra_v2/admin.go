package extrav2

import (
	"context"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

// AdminV2 struct
type AdminV2 struct {
	client           influxdb2.Client
	organizationName string
	bucketName       string
	ctx              context.Context
}

// NewAdminClient returns influxdb admin client
func NewAdminClient(ctx context.Context, client influxdb2.Client, organizationName, bucketName string) *AdminV2 {
	return &AdminV2{ctx: ctx, client: client, organizationName: organizationName, bucketName: bucketName}
}

// IsBucketAvailable returns the availability of the database
func (av2 *AdminV2) IsBucketAvailable() (bool, error) {
	bucketDomain, err := av2.client.BucketsAPI().FindBucketByName(av2.ctx, av2.bucketName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return bucketDomain.Name == av2.bucketName, nil
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

	// get organization ID
	orgDomain, err := av2.client.OrganizationsAPI().FindOrganizationByName(av2.ctx, av2.organizationName)
	if err != nil {
		return err
	}
	_, err = av2.client.BucketsAPI().CreateBucketWithName(av2.ctx, orgDomain, av2.bucketName)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	zap.L().Info("metrics bucket created", zap.String("organizationName", av2.organizationName), zap.String("bucketName", av2.bucketName))
	return nil
}
