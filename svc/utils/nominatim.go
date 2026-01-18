package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	UrlPrefix string
}

func NewClient(urlPrefix string) *Client {
	return &Client{UrlPrefix: fmt.Sprintf("http://%s", urlPrefix)}
}

func (c *Client) Reverse(lat, lon float64) (map[string]interface{}, error) {
	var r map[string]interface{}
	params := fmt.Sprintf("lat=%v&lon=%v&format=geojson&zoom=18", lat, lon)
	client := http.DefaultClient
	resp, err := client.Get(c.UrlPrefix + "/reverse?" + params)
	if err != nil {
		return r, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) Search(name string, longs, lats []float64) (map[string]interface{}, error) {
	var r map[string]interface{}
	viewBox := fmt.Sprintf(`%v,%v,%v,%v`, longs[0], lats[0], longs[1], lats[1])
	params := fmt.Sprintf("q=%s&viewbox=%s&bounded=1&limit=6&format=geojson&zoom=18", url.QueryEscape(name), viewBox)
	client := http.DefaultClient

	resp, err := client.Get(c.UrlPrefix + "/search?" + params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("nominatim request failed with code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
