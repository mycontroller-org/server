package api

import (
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/version"
)

func (c *Client) GetServerVersion() (*version.Version, error) {
	res, err := c.executeJson(API_VERSION, http.MethodGet, nil, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}

	ver := &version.Version{}
	err = json.Unmarshal(res.Body, ver)
	if err != nil {
		return nil, err
	}
	return ver, nil
}
