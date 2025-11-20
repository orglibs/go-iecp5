// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
)

// DefaultReconnectInterval defined default value
const DefaultReconnectInterval = 1 * time.Minute

type seqPending struct {
	seq         uint16
	sendTime    time.Time
	asduPending []byte
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

func receiveLoop(conn net.Conn, rcvRaw chan []byte) {
	for {
		rawData := make([]byte, APDUSizeMax)
		for rdCnt, length := 0, 2; rdCnt < length; {
			byteCount, err := io.ReadFull(conn, rawData[rdCnt:length])
			if err != nil {
				// See: https://github.com/golang/go/issues/4373
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe) ||
					strings.Contains(err.Error(), "use of closed network connection") {
					slog.Error("receive failed", "error", err)

					return
				}

				//nolint:errorlint
				if e, ok := err.(net.Error); ok && !e.Timeout() {
					slog.Error("receive failed", "error", err)

					return
				}

				if rdCnt == 0 && err == io.EOF {
					slog.Error("remote connect closed", "error", err)

					return
				}
			}

			rdCnt += byteCount
			switch rdCnt {
			case 0:
				continue
			case 1:
				if rawData[0] != StartFrame {
					rdCnt = 0

					continue
				}
			default:
				if rawData[0] != StartFrame {
					rdCnt, length = 0, 2

					continue
				}

				length = int(rawData[1]) + 2
				if length < APCICtlFiledSize+2 || length > APDUSizeMax {
					rdCnt, length = 0, 2

					continue
				}

				if rdCnt == length {
					apdu := rawData[:length]
					slog.Debug("RX Raw", "rx", apdu)
					rcvRaw <- apdu
				}
			}
		}
	}
}
