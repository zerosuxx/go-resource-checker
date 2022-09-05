package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
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
	ip, err := lookupIPv4WithRetry(u.Hostname(), 5, 3)

	if err != nil {
		return err
	}

	u.Host = ip.String() + ":" + u.Port()
	log.Println("Resolved IP address: " + u.Hostname())

	dialer := net.Dialer{
		Timeout: timeout,
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

func lookupIPv4WithRetry(hostName string, tryCount int, waitTimeInSeconds int) (net.IP, error) {
	var ip net.IP
	var err error
	for i := 1; i <= tryCount; i++ {
		var ips, err = net.DefaultResolver.LookupIP(context.Background(), "ip4", hostName)

		if err != nil {
			time.Sleep(time.Duration(waitTimeInSeconds) * time.Second)
			continue
		}

		ip = ips[0]
	}

	return ip, err
}
