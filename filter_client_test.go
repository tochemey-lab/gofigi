/*
 * MIT License
 *
 * Copyright (c) 2024 tochemey
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package gofigi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"

	"github.com/tochemey/gofigi/model"
)

type filterClientSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFilterClient(t *testing.T) {
	suite.Run(t, new(filterClientSuite))
}

func (s *filterClientSuite) TestFilter() {
	s.Run("with happy path", func() {
		// create the context
		ctx := context.TODO()
		// create a request ID
		requestID := uuid.NewString()
		// create the mock test server to mock
		testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			const fileName = "testdata/filter-resp.json"
			// validate the request url to make sure we are hitting the right endpoint
			path := request.URL.Path
			if path != filterResourcePath {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// set the response headers
			writer.Header().Add("content-type", "application/json; charset=utf-8")
			writer.Header().Add("X-Request-Id", requestID)
			writer.WriteHeader(http.StatusOK)
			// let us read the test data file
			// this line will panic since we don't handle the error which is ok
			file, _ := os.ReadFile(fileName)
			// write the response body and panic in case there is an error
			_, _ = writer.Write(file)
		}))
		s.Assert().NotNil(testServer)
		// close the test server once done with the test
		defer testServer.Close()
		// create the mock endpoint and auth key
		mockAPIKey := uuid.NewString()
		// mock the configuration
		// create the request
		req := &model.FilterRequest{
			SecurityType:            "Common Stock",
			MarketSecDes:            model.EquityMarketSector,
			IncludeUnlistedEquities: false,
			MicCode:                 "XNYS",
		}
		// create expected token
		expectedNext := "QW9Fc1FrSkhNREF3UWtJd01UZ3ggMQ==.ikDNdboMpp/GFctV8PIxEmmI2w7Kz1kSqw9QwPG7gZo="
		// create the client
		filterClient := &FilterClient{
			client{
				httpClient: http.DefaultClient,
				limiter:    rate.NewLimiter(rate.Every(1*time.Minute/20), 20),
				apiKey:     mockAPIKey,
				baseURL:    testServer.URL,
			},
		}
		// make a search
		resp, err := filterClient.Filter(ctx, req)
		// run some assertions
		s.Assert().NoError(err)
		s.Assert().NotNil(resp)
		s.Assert().NotEmpty(resp.Data)
		s.Assert().Equal(34636, resp.Total)
		s.Assert().Empty(resp.Error)
		s.Assert().NotEmpty(resp.Next)
		s.Assert().Equal(expectedNext, resp.Next)
	})
	s.Run("with not OK status", func() {
		// create the context
		ctx := context.TODO()
		// create a request ID
		requestID := uuid.NewString()
		// create the mock test server to mock
		testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// validate the request url to make sure we are hitting the right endpoint
			path := request.URL.Path
			if path != filterResourcePath {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// set the response headers
			writer.Header().Add("content-type", "application/json; charset=utf-8")
			writer.Header().Add("X-Request-Id", requestID)
			writer.WriteHeader(http.StatusServiceUnavailable)
		}))
		s.Assert().NotNil(testServer)
		// close the test server once done with the test
		defer testServer.Close()
		// create the mock endpoint and auth key
		mockAPIKey := uuid.NewString()
		// mock the configuration

		// create the request
		req := &model.FilterRequest{
			SecurityType:            "Common Stock",
			MarketSecDes:            model.EquityMarketSector,
			IncludeUnlistedEquities: false,
			MicCode:                 "XNYS",
		}
		// create the client
		client := &FilterClient{
			client{
				httpClient: http.DefaultClient,
				limiter:    rate.NewLimiter(rate.Every(1*time.Minute/20), 20),
				apiKey:     mockAPIKey,
				baseURL:    testServer.URL,
			},
		}
		// make a search
		resp, err := client.Filter(ctx, req)
		s.Assert().Error(err)
		s.Assert().EqualError(err, fmt.Sprintf("ErrValidator: response error for %s%s: unexpected status: %v", testServer.URL, filterResourcePath, http.StatusServiceUnavailable))
		s.Assert().Nil(resp)
	})
	s.Run("with unexpected json response", func() {
		// create the context
		ctx := context.TODO()
		// create a request ID
		requestID := uuid.NewString()
		// create the mock test server to mock
		testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// validate the request url to make sure we are hitting the right endpoint
			path := request.URL.Path
			if path != filterResourcePath {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// set the response headers
			writer.Header().Add("content-type", "application/json; charset=utf-8")
			writer.Header().Add("X-Request-Id", requestID)
			writer.WriteHeader(http.StatusOK)
			// write the response body and panic in case there is an error
			_, _ = writer.Write([]byte(`{"name": "unmatched"}`))
		}))
		s.Assert().NotNil(testServer)
		// close the test server once done with the test
		defer testServer.Close()
		// create the mock endpoint and auth key
		mockAPIKey := uuid.NewString()
		// create the request
		req := &model.FilterRequest{
			SecurityType:            "Common Stock",
			MarketSecDes:            model.EquityMarketSector,
			IncludeUnlistedEquities: false,
			MicCode:                 "XNYS",
		}
		// create the client
		client := &FilterClient{
			client{
				httpClient: http.DefaultClient,
				limiter:    rate.NewLimiter(rate.Every(1*time.Minute/20), 20),
				apiKey:     mockAPIKey,
				baseURL:    testServer.URL,
			},
		}
		// make a search
		resp, err := client.Filter(ctx, req)
		s.Assert().NoError(err)
		s.Assert().NotNil(resp)
		s.Assert().Empty(resp.Data)
		s.Assert().Equal(0, resp.Total)
		s.Assert().Empty(resp.Error)
		s.Assert().Empty(resp.Next)
	})
	s.Run("with invalid json response", func() {
		// create the context
		ctx := context.TODO()
		// create a request ID
		requestID := uuid.NewString()
		// create the mock test server to mock
		testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// validate the request url to make sure we are hitting the right endpoint
			path := request.URL.Path
			if path != filterResourcePath {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// set the response headers
			writer.Header().Add("content-type", "application/json; charset=utf-8")
			writer.Header().Add("X-Request-Id", requestID)
			writer.WriteHeader(http.StatusOK)
			// write the response body and panic in case there is an error
			_, _ = writer.Write([]byte(`["Hello", 3.14, true, ]`))
		}))
		s.Assert().NotNil(testServer)
		// close the test server once done with the test
		defer testServer.Close()
		// create the mock endpoint and auth key
		mockAPIKey := uuid.NewString()
		// create the request
		req := &model.FilterRequest{
			SecurityType:            "Common Stock",
			MarketSecDes:            model.EquityMarketSector,
			IncludeUnlistedEquities: false,
			MicCode:                 "XNYS",
		}
		client := &FilterClient{
			client{
				httpClient: http.DefaultClient,
				limiter:    rate.NewLimiter(rate.Every(1*time.Minute/20), 20),
				apiKey:     mockAPIKey,
				baseURL:    testServer.URL,
			},
		}
		// make a search
		resp, err := client.Filter(ctx, req)
		s.Assert().Error(err)
		s.Assert().Nil(resp)
	})
	s.Run("with invalid API Key", func() {
		// create the context
		ctx := context.TODO()
		// create a request ID
		requestID := uuid.NewString()
		// create the mock test server to mock
		testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// validate the request url to make sure we are hitting the right endpoint
			path := request.URL.Path
			if path != filterResourcePath {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			// set the response headers
			writer.Header().Add("content-type", "application/json; charset=utf-8")
			writer.Header().Add("X-Request-Id", requestID)
			// define valid API key
			validAPIKey := uuid.NewString()
			apiKey := request.Header.Get("X-OPENFIGI-APIKEY")
			if apiKey != validAPIKey {
				writer.WriteHeader(http.StatusUnauthorized)
				// write the response body and panic in case there is an error
				_, _ = writer.Write([]byte(`Invalid API key.`))
				return
			}
		}))
		s.Assert().NotNil(testServer)
		// close the test server once done with the test
		defer testServer.Close()
		// create the mock endpoint and auth key
		mockAPIKey := uuid.NewString()

		// create the request
		req := &model.FilterRequest{
			SecurityType:            "Common Stock",
			MarketSecDes:            model.EquityMarketSector,
			IncludeUnlistedEquities: false,
			MicCode:                 "XNYS",
		}
		// create the client
		client := &FilterClient{
			client{
				httpClient: http.DefaultClient,
				limiter:    rate.NewLimiter(rate.Every(1*time.Minute/20), 20),
				apiKey:     mockAPIKey,
				baseURL:    testServer.URL,
			},
		}
		// make a search
		resp, err := client.Filter(ctx, req)
		s.Assert().Error(err)
		s.Assert().EqualError(err, fmt.Sprintf("ErrValidator: response error for %s%s: unexpected status: %v", testServer.URL, filterResourcePath, http.StatusUnauthorized))
		s.Assert().Nil(resp)
	})
}
