package proxy

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/events"
)

// HandlerFunc is the function interface for proxying API Gateway events to standard http.Handler
type HandlerFunc func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

// Handler converts the APIGatewayV2HTTPRequest into a standard golang http.Request and provides a simple writer.
func Handler(handler http.Handler) HandlerFunc {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		req, err := RequestForEvent(event)
		if err != nil {
			return InternalServerErrorResponse(), err
		}

		w := NewResponseWriter().(*writer)

		handler.ServeHTTP(w, req.WithContext(ctx))

		return w.response()
	}
}

// InternalServerErrorResponse returns a standard 500 error response
func InternalServerErrorResponse() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 500,
		MultiValueHeaders: map[string][]string{
			`Content-Type`: {`application/json`},
		},
		Body:            `{"errorCode":"500"}`,
		IsBase64Encoded: false,
	}
}

// RequestForEvent converts the APIGatewayV2HTTPRequest into a standard golang http.Request
func RequestForEvent(e events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	body := []byte(e.Body)
	if e.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return nil, err
		}
		body = decoded
	}

	uri := &url.URL{}

	path := e.RawPath
	if len(path) == 0 {
		path = e.RequestContext.HTTP.Path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	uri.Scheme = `https`
	uri.Path = path
	uri.Host = e.RequestContext.DomainName

	if len(uri.Host) == 0 {
		uri.Host = `localhost`
	}
	if port, ok := e.Headers[`X-Forwarded-Port`]; ok {
		uri.Host = net.JoinHostPort(uri.Host, port)
	}

	if len(e.RawQueryString) > 0 {
		uri.RawQuery = e.RawQueryString
	} else if len(e.QueryStringParameters) > 0 {
		query := url.Values{}

		for k, v := range e.QueryStringParameters {
			query.Add(k, v)
		}

		uri.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(strings.ToUpper(e.RequestContext.HTTP.Method), uri.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range e.Headers {
		for _, val := range strings.Split(v, ",") {
			if k == `User-Agent` {
				req.Header.Set(k, v)
			} else {
				req.Header.Add(k, strings.Trim(val, " "))
			}
		}
	}

	req.RemoteAddr = e.RequestContext.HTTP.SourceIP
	req.RequestURI = uri.String()

	return req, nil
}

// NewResponseWriter returns a new ResponseWriter for the request
func NewResponseWriter() http.ResponseWriter {
	return &writer{
		status:  defaultStatusCode,
		headers: make(http.Header),
	}
}

const defaultStatusCode = -1
const contentTypeHeaderKey = "Content-Type"

type writer struct {
	status  int
	headers http.Header
	body    bytes.Buffer
}

func (w *writer) Header() http.Header {
	return w.headers
}

func (w *writer) Write(b []byte) (int, error) {
	if w.status == defaultStatusCode {
		w.status = http.StatusOK
	}

	if w.Header().Get(contentTypeHeaderKey) == "" {
		w.Header().Add(contentTypeHeaderKey, http.DetectContentType(b))
	}

	return w.body.Write(b)
}

func (w *writer) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *writer) response() (events.APIGatewayV2HTTPResponse, error) {
	if w.status == defaultStatusCode {
		return InternalServerErrorResponse(), errors.New("response status code not set")
	}

	res := events.APIGatewayV2HTTPResponse{
		StatusCode:        w.status,
		MultiValueHeaders: w.headers,
	}

	bodyBytes := w.body.Bytes()
	if utf8.Valid(bodyBytes) {
		res.Body = string(bodyBytes)
		res.IsBase64Encoded = false
	} else {
		res.Body = base64.StdEncoding.EncodeToString(bodyBytes)
		res.IsBase64Encoded = true
	}

	return res, nil
}

var _ http.ResponseWriter = new(writer)
