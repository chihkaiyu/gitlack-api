package resource

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type Util interface {
	Request(string, string, map[string]string, map[string]string, map[string]string) (*http.Response, error)
}

// Client stores token for getting user, project resources
type external struct {
	client *http.Client
}

func NewUtil() Util {
	client := &http.Client{}
	return &external{
		client: client,
	}
}

func (e *external) Request(method, endpoint string, header map[string]string, param map[string]string, body map[string]string) (*http.Response, error) {
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

	res, err := e.client.Do(req)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return res, nil
}
