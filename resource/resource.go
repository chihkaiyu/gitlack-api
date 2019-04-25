package resource

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Client handles HTTP request
type Client interface {
	Get(string, map[string]string, map[string]string, map[string]string) (*http.Response, error)
	Post(string, map[string]string, map[string]string, map[string]string) (*http.Response, error)
}

// Client hold a HTTP client
type client struct {
	client *http.Client
}

// NewClient return a Request which handles HTTP request
func NewClient() Client {
	return &client{
		client: &http.Client{},
	}
}

func (c *client) Get(endpoint string, header map[string]string, param map[string]string, body map[string]string) (*http.Response, error) {
	req, err := prepareRequest(http.MethodGet, endpoint, header, param, body)
	response, err := c.client.Do(req)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return response, nil
}

func (c *client) Post(endpoint string, header map[string]string, param map[string]string, body map[string]string) (*http.Response, error) {
	req, err := prepareRequest(http.MethodPost, endpoint, header, param, body)
	response, err := c.client.Do(req)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return response, nil
}

func prepareRequest(method, endpoint string, header map[string]string, param map[string]string, body map[string]string) (*http.Request, error) {
	// add body
	reqBody := url.Values{}
	for k, v := range body {
		reqBody.Add(k, v)
	}

	req, err := http.NewRequest(method, endpoint, strings.NewReader(reqBody.Encode()))
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	// add header
	for k, v := range header {
		req.Header.Add(k, v)
	}

	// add query string parameters
	params := req.URL.Query()
	for k, v := range param {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()

	return req, nil
}
