package diagnostics

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opencensus.io/stats/view"
)

func TestFastHTTPMiddleware(t *testing.T) {
	requestBody := "fake_requestDaprBody"
	responseBody := "fake_responseDaprBody"

	testRequestCtx := fakeFastHTTPRequestCtx(requestBody)

	fakeHandler := func(ctx *fasthttp.RequestCtx) {
		time.Sleep(100 * time.Millisecond)
		ctx.Response.SetBodyRaw([]byte(responseBody))
	}

	// create test httpMetrics
	testHTTP := newHTTPMetrics()
	testHTTP.Init("fakeID")

	handler := testHTTP.FastHTTPMiddleware(fakeHandler)

	// act
	handler(testRequestCtx)

	// assert
	rows, err := view.RetrieveData("http/server/request_count")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, "app_id", rows[0].Tags[0].Key.Name())
	assert.Equal(t, "fakeID", rows[0].Tags[0].Value)
	assert.Equal(t, "method", rows[0].Tags[1].Key.Name())
	assert.Equal(t, "POST", rows[0].Tags[1].Value)
	assert.Equal(t, "path", rows[0].Tags[2].Key.Name())
	assert.Equal(t, "/invoke/method/testmethod", rows[0].Tags[2].Value)

	rows, err = view.RetrieveData("http/server/request_bytes")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, "app_id", rows[0].Tags[0].Key.Name())
	assert.Equal(t, "fakeID", rows[0].Tags[0].Value)
	assert.True(t, (rows[0].Data).(*view.DistributionData).Min == float64(len([]byte(requestBody))))

	rows, err = view.RetrieveData("http/server/response_bytes")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.True(t, (rows[0].Data).(*view.DistributionData).Min == float64(len([]byte(responseBody))))

	rows, err = view.RetrieveData("http/server/latency")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.True(t, (rows[0].Data).(*view.DistributionData).Min >= 100.0)
}

func fakeFastHTTPRequestCtx(expectedBody string) *fasthttp.RequestCtx {
	expectedMethod := fasthttp.MethodPost
	expectedRequestURI := "/invoke/method/testmethod"
	expectedTransferEncoding := "encoding"
	expectedHost := "dapr.io"
	expectedRemoteAddr := "1.2.3.4:6789"
	expectedHeader := map[string]string{
		"Correlation-ID":  "e6f4bb20-96c0-426a-9e3d-991ba16a3ebb",
		"XXX-Remote-Addr": "192.168.0.100",
	}

	var ctx fasthttp.RequestCtx
	var req fasthttp.Request

	req.Header.SetMethod(expectedMethod)
	req.SetRequestURI(expectedRequestURI)
	req.Header.SetHost(expectedHost)
	req.Header.Add(fasthttp.HeaderTransferEncoding, expectedTransferEncoding)
	req.Header.SetContentLength(len([]byte(expectedBody)))
	req.BodyWriter().Write([]byte(expectedBody)) // nolint:errcheck

	for k, v := range expectedHeader {
		req.Header.Set(k, v)
	}

	remoteAddr, _ := net.ResolveTCPAddr("tcp", expectedRemoteAddr)

	ctx.Init(&req, remoteAddr, nil)

	return &ctx
}
