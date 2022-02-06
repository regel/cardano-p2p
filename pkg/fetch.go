package pkg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func buildUrl(endpoint string, magic int64, max int64, ipv int64) string {
	base, err := url.Parse(endpoint)
	if err != nil {
		panic("Can't parse endpoint base url")
	}
	values := url.Values{
		"magic": []string{strconv.FormatInt(magic, 10)},
		"max":   []string{strconv.FormatInt(max, 10)},
		"ipv":   []string{strconv.FormatInt(ipv, 10)},
	}
	relative := &url.URL{
		Path:     "/htopology/v1/fetch/",
		RawQuery: values.Encode(),
	}

	return base.ResolveReference(relative).String()
}

func Fetch(bg context.Context, clioEndpoint string, magic int64, fetchMax int64, ipVersion int64) ([]byte, error) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var client = &http.Client{
		Transport: netTransport,
	}

	ctx, cancel := context.WithTimeout(bg, requestMaxWaitTime)
	defer cancel()

	url := buildUrl(clioEndpoint, magic, fetchMax, ipVersion)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create request: %v", err)
	}
	req.Close = true
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if e, ok := err.(net.Error); ok && e.Timeout() {
		return nil, fmt.Errorf("Get '%s' timeout", url)
	} else if err != nil {
		return nil, fmt.Errorf("Cannot do request: %v", err)
	}
	buf, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxResponseLen))
	if err != nil {
		return nil, fmt.Errorf("Cannot get all response body at url '%s': %v", url, err)
	}
	return buf, nil
}
