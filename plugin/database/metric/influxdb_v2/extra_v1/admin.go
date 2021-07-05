package extrav1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	httpclient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	"go.uber.org/zap"
)

// QueryV1 struct
type AdminV1 struct {
	client  *httpclient.Client
	headers map[string]string
	url     string
	bucket  string
}

// NewAdminClient returns admin client
func NewAdminClient(uri string, insecureSkipVerify bool, bucket, username, password string) *AdminV1 {
	headers, newClient := newClient(uri, insecureSkipVerify, username, password)

	return &AdminV1{
		client:  newClient,
		url:     fmt.Sprintf("%s/query", uri),
		headers: headers,
		bucket:  bucket,
	}
}

// CreateBucket adds a database to influxdb
func (av1 *AdminV1) CreateBucket() error {
	queryParams := map[string]interface{}{
		"q": fmt.Sprintf("CREATE DATABASE \"%s\"", av1.bucket),
	}
	_, responseBody, err := av1.client.Request(av1.url, http.MethodGet, av1.headers, queryParams, nil, http.StatusOK)
	if err != nil {
		zap.L().Error("error on calling api", zap.Error(err))
		return err
	}

	queryResult := QueryResult{}
	err = json.Unmarshal(responseBody, &queryResult)
	if err != nil {
		return err
	}

	if queryResult.Error != "" {
		return errors.New(queryResult.Error)
	}

	zap.L().Info("metrics database available or created", zap.String("database", av1.bucket))
	return nil
}
