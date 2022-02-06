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

func buildPushUrl(endpoint string, magic int64, port int64, blockNo int64) string {
	base, err := url.Parse(endpoint)
	if err != nil {
		panic("Can't parse endpoint base url")
	}
	values := url.Values{
		"magic":   []string{strconv.FormatInt(magic, 10)},
		"port":    []string{strconv.FormatInt(port, 10)},
		"blockNo": []string{strconv.FormatInt(blockNo, 10)},
	}
	relative := &url.URL{
		Path:     "/htopology/v1/",
		RawQuery: values.Encode(),
	}

	return base.ResolveReference(relative).String()
}

func PushBlockNo(bg context.Context, clioEndpoint string, magic int64, port int64, blockNo int64) ([]byte, error) {
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

	url := buildPushUrl(clioEndpoint, magic, port, blockNo)
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
