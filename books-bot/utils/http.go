package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func GetRequest(requestURL string, headers map[string]string) ([]byte, error) {
	res, err := GetRequestRaw(requestURL, headers)
	if err != nil {
		return []byte{}, err
	}
	content, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("cannot load %s: got status code %d: %s", requestURL, res.StatusCode, string(content))
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read the main sitemap contents at %s: %v", requestURL, err)
	}
	return content, nil
}

func GetRequestRaw(requestURL string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request for the url %s: %v", requestURL, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute request for the url %s: %v", requestURL, err)
	}

	return res, nil
}

func post(requestURL string, data []byte, headers http.Header, client *http.Client) ([]byte, error) {
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("cannot create request for the url %s: %v", requestURL, err)
	}
	req.Header = headers
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute request for the url %s: %v", requestURL, err)
	}
	content, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	return content, err
}
