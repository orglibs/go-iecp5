// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
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
	return openConnectionContext(context.Background(), uri, tlsc, timeout)
}

func openConnectionContext(ctx context.Context, uri *url.URL, tlsc *tls.Config, timeout time.Duration) (net.Conn, error) {
	switch uri.Scheme {
	case "tcp":
		con, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", uri.Host)
		if err != nil {
			return nil, fmt.Errorf("tcp error:%w", err)
		}

		return con, nil
	case "ssl":
		fallthrough
	case "tls":
		fallthrough
	case "tcps":
		cons, err := (&tls.Dialer{NetDialer: &net.Dialer{Timeout: timeout}, Config: tlsc}).DialContext(ctx, "tcp", uri.Host)
		if err != nil {
			return nil, fmt.Errorf("tcps error:%w", err)
		}

		return cons, nil
	}

	return nil, errors.New("unknown protocol")
}

func receiveLoop(conn net.Conn, rcvRaw chan []byte) {
	receiveLoopContext(context.Background(), conn, rcvRaw)
}

// receiveLoopContext 在连接关闭、截断帧或取消时退出；不对 EOF/ClosedPipe 空转。
// 使用可取消的队列发送，避免接收队列满时阻塞主站关闭。
func receiveLoopContext(ctx context.Context, conn net.Conn, rcvRaw chan []byte) {
	for {
		rawData := make([]byte, APDUSizeMax)
		for rdCnt, length := 0, 2; rdCnt < length; {
			byteCount, err := io.ReadFull(conn, rawData[rdCnt:length])
			if err != nil {
				if ctx.Err() == nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.ErrClosedPipe) {
					slog.Debug("receive stopped", "error", err)
				}
				return
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
					select {
					case rcvRaw <- apdu:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}
