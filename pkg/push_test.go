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

const samplePushOkResponse = `
{
  "resultcode": "201",
  "datetime": "2022-02-06 19:14:49",
  "clientIp": "127.0.0.1",
  "iptype": 4,
  "msg": "nice to meet you"
}`

// 200 OK with body:
const samplePushKoResponse = `
{
  "resultcode": "503",
  "datetime": "2022-02-06 19:21:15",
  "clientIp": "127.0.0.1",
  "msg": "blockNo 3195424 seems out of sync. please retry"
}`

func TestPushOkResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/" {
			rsp = samplePushOkResponse
			w.Header()["content-type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := PushBlockNo(context, ts.URL, 1, 6001, 10000)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, samplePushOkResponse, strings.TrimSuffix(string(out), "\n"))
}

func TestPushKoResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/" {
			rsp = samplePushKoResponse
			w.Header()["content-type"] = []string{"application/json"}
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := PushBlockNo(context, ts.URL, 1, 6001, 10000)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, samplePushKoResponse, strings.TrimSuffix(string(out), "\n"))
}

func TestPushKo400Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/htopology/v1/" {
			rsp = samplePushKoResponse
			w.Header()["content-type"] = []string{"application/json"}
			w.WriteHeader(400)
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	context := context.Background()
	out, err := PushBlockNo(context, ts.URL, 1, 6001, 10000)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.EqualValues(t, samplePushKoResponse, strings.TrimSuffix(string(out), "\n"))
}
