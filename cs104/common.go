// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"
)

// DefaultReconnectInterval defined default value
const DefaultReconnectInterval = 1 * time.Minute

type seqPending struct {
	seq      uint16
	sendTime time.Time
}

func openConnection(uri *url.URL, tlsc *tls.Config, timeout time.Duration) (net.Conn, error) {
	switch uri.Scheme {
	case "tcp":
		con, err := net.DialTimeout("tcp", uri.Host, timeout)
		if err != nil {
			return nil, fmt.Errorf("tcp error:%w", err)
		}
		return con, nil
	case "ssl":
		fallthrough
	case "tls":
		fallthrough
	case "tcps":
		cons, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", uri.Host, tlsc)
		if err != nil {
			return nil, fmt.Errorf("tcps error:%w", err)
		}
		return cons, nil
	}
	return nil, errors.New("unknown protocol")
}
