// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package clog

import (
	"io"
	"log"
	"os"
	"path"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFile  = "clog.log"
	pathLogs = ""
)

// NewLogger Sets the logger options
func NewLogger() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   path.Join(pathLogs, logFile),
		MaxSize:    1,
		MaxBackups: 4,
		MaxAge:     365,
		Compress:   false,
		LocalTime:  true,
	}))
}
