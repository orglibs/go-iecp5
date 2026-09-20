// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"errors"
)

// errors definition
var (
	ErrUseClosedConnection = errors.New("use of closed connection")
	// ErrInvalidAPCI 表示控制域不符合 IEC104，调用方应丢弃整帧，不改变状态。
	ErrInvalidAPCI = errors.New("invalid APCI")
	ErrBufferFull  = errors.New("buffer is full")
	ErrNotActive   = errors.New("server is not active")
)
