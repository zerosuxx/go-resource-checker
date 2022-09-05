package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ResourceChecker struct {
	CheckSuccessOnHealthCheck bool
}

func (c ResourceChecker) Check(u *url.URL, timeout int) error {
	timeoutDuration := time.Second * time.Duration(timeout)

	if u.Scheme == "tcp" || u.Scheme == "udp" {
		return c.checkNetwork(u, timeoutDuration)
	} else {
		return c.checkHTTP(u, timeoutDuration)
	}
}

func (c ResourceChecker) checkNetwork(u *url.URL, timeout time.Duration) error {
	dialer := net.Dialer{
		Timeout: timeout,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, network, "8.8.8.8:53")
			},
		},
	}
	connection, err := dialer.Dial(u.Scheme, u.Host)

	if err != nil {
		return err
	}

	var closeError error
	defer func() {
		closeError = connection.Close()
	}()

	return closeError
}

func (c ResourceChecker) checkHTTP(u *url.URL, timeout time.Duration) error {
	client := http.Client{Timeout: timeout}
	response, err := client.Get(u.String())
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode > 399 {
		return errors.New(u.String() + " is unavailable! [" + strconv.Itoa(response.StatusCode) + "]")
	}

	if c.CheckSuccessOnHealthCheck && strings.HasSuffix(u.String(), "/healthcheck") {
		body := getBytesFromBody(response.Body)
		response.Body = getBodyFromBytes(body)

		successResponse := JsonResponse{}
		_ = json.Unmarshal(body, &successResponse)

		if !successResponse.Success {
			return errors.New(u.String() + " is not healthy!")
		}
	}

	return nil
}

func getBytesFromBody(body io.ReadCloser) []byte {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = ioutil.ReadAll(body)
	}

	return bodyBytes
}

func getBodyFromBytes(data []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(data))
}
