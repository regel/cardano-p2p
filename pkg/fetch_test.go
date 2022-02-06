package pkg

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleFetchOkResponse = `
{
  "resultcode": "200",
  "datetime": "2022-02-05 16:54:21",
  "clientIp": "127.0.0.1",
  "iptype": 4,
  "msg": "welcome to the topology",
  "Producers": [
    {
      "addr": "23.94.134.119",
      "port": 5001,
      "valency": 1
    }
  ]
}`

// 200 OK with body:
const sampleFetchKoResponse = `
{
  "resultcode": "402",
  "datetime":"2022-02-05 16:59:00",
  "clientIp": "127.0.0.1",
  "iptype": 6,
  "msg": "IP is not (yet) allowed to fetch this list",
  "Producers": [ { "addr": "relays-new.cardano-mainnet.iohk.io", "port": 3001, "valency": 2 } ]
}`

func TestFetchOkResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/fetch/" {
			rsp = sampleFetchOkResponse
			w.Header()["content-type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := Fetch(context, ts.URL, 1, 1, 4)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, sampleFetchOkResponse, strings.TrimSuffix(string(out), "\n"))
}

func TestFetchKoResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/fetch/" {
			rsp = sampleFetchKoResponse
			w.Header()["content-type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := Fetch(context, ts.URL, 1, 1, 4)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, sampleFetchKoResponse, strings.TrimSuffix(string(out), "\n"))
}

func TestFetchKo400Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/fetch/" {
			rsp = sampleFetchKoResponse
			w.Header()["content-type"] = []string{"application/json"}
			w.WriteHeader(400)
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := Fetch(context, ts.URL, 1, 1, 4)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, sampleFetchKoResponse, strings.TrimSuffix(string(out), "\n"))
}
