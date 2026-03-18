/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package common

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type HTTPMethod string

const (
	MethodGet    HTTPMethod = "GET"
	MethodPost   HTTPMethod = "POST"
	MethodPut    HTTPMethod = "PUT"
	MethodDelete HTTPMethod = "DELETE"
)

const (
	RetryCount  = 3
	maxBodySize = 1 << 20
)

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true
	}

	return false
}

func drainBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

type retryableTransport struct {
	transport http.RoundTripper
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil && req.GetBody == nil {
		if req.ContentLength >= 0 && req.ContentLength > maxBodySize {
			return nil, fmt.Errorf("request body too large to buffer for retry (size %d exceeds max allowed %d)", req.ContentLength, maxBodySize)
		}

		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(bodyBytes)), nil
		}
	}

	resp, err := t.transport.RoundTrip(req)
	retries := 0
	for shouldRetry(err, resp) && retries < RetryCount {
		time.Sleep(backoff(retries))

		drainBody(resp)

		if req.Body != nil {
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				req.Body = newBody
			}
		}

		resp, err = t.transport.RoundTrip(req)
		retries++
	}

	return resp, err
}

func NewRetryableTransport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &retryableTransport{transport: rt}
}
