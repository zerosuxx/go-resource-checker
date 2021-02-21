package checker

import (
	"net"
	"net/url"
	"time"
)

type ResourceChecker struct {
}

func (checker ResourceChecker) Check(u *url.URL, timeout int) error {
	dialer := net.Dialer{Timeout: time.Second * time.Duration(timeout)}

	connection, err := dialer.Dial(u.Scheme, u.Host)
	if err != nil {
		return err
	}

	defer func() {
		err = connection.Close()
	}()

	return err
}
