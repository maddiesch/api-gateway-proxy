package proxy_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	proxy "github.com/maddiesch/api-gateway-proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleEventV2 = `
{
  "version": "2.0",
  "routeKey": "GET /\u003cpath:path\u003e",
  "rawPath": "/foo/bar",
  "rawQueryString": "baz=1",
  "headers": {
    "Connection": "close",
    "Host": "127.0.0.1:3000",
    "User-Agent": "Paw/3.2 (Macintosh; OS X/10.16.0) GCDHTTPRequest",
    "X-Forwarded-Port": "3000",
    "X-Forwarded-Proto": "http"
  },
  "queryStringParameters": {
    "baz": "1"
  },
  "pathParameters": {
    "path": "foo/bar"
  },
  "requestContext": {
    "routeKey": "GET /\u003cpath:path\u003e",
    "accountId": "123456789012",
    "stage": "",
    "requestId": "4036e453-d4db-47db-8d34-97bf81af40dd",
    "apiId": "1234567890",
    "domainName": "",
    "domainPrefix": "",
    "time": "",
    "timeEpoch": 0,
    "http": {
      "method": "GET",
      "path": "/foo/bar",
      "protocol": "HTTP/1.1",
      "sourceIp": "127.0.0.1",
      "userAgent": "Custom User Agent String"
    }
  },
  "isBase64Encoded": false
}
`

func TestRequestForEvent(t *testing.T) {
	t.Run("given a valid request", func(t *testing.T) {
		event := events.APIGatewayV2HTTPRequest{}

		err := json.Unmarshal([]byte(exampleEventV2), &event)

		require.NoError(t, err)

		t.Run("it does not return an error", func(t *testing.T) {
			_, err := proxy.RequestForEvent(event)

			assert.NoError(t, err)
		})

		t.Run("it returns a non-nil request", func(t *testing.T) {
			req, _ := proxy.RequestForEvent(event)

			assert.NotNil(t, req)
		})
	})
}
