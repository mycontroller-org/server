package api

import (
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
)

func (c *Client) Login(username, password, token, expiresIn string) (*handlerTY.JwtTokenResponse, error) {
	req := &handlerTY.UserLogin{
		Username:  username,
		Password:  password,
		SvcToken:  token,
		ExpiresIn: expiresIn,
	}
	res, err := c.executeJson(API_LOGIN, http.MethodPost, nil, nil, req, http.StatusOK)
	if err != nil {
		return nil, err
	}

	result := &handlerTY.JwtTokenResponse{}
	err = json.Unmarshal(res.Body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
