// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"errors"
	"time"
)

const (
	// Port is the IANA registered port number for unsecure connection.
	Port = 2404

	// PortSecure is the IANA registered port number for secure connection.
	PortSecure = 19998
)

// defines an IEC 60870-5-104 configuration range
const (
	// "t0" range[1, 255]s default value 30s
	ConnectTimeout0Min = 1 * time.Second
	ConnectTimeout0Max = 255 * time.Second

	// "t1" range[1, 255]s default value 15s. See IEC 60870-5-104, figure 18.
	SendUnAckTimeout1Min = 1 * time.Second
	SendUnAckTimeout1Max = 255 * time.Second

	// "t2" range[1, 255]s default value 10s, See IEC 60870-5-104, figure 10.
	RecvUnAckTimeout2Min = 1 * time.Second
	RecvUnAckTimeout2Max = 255 * time.Second

	// "t3" range[1 second, 48 hours] default value 20 s, See IEC 60870-5-104, subclass 5.2.
	IdleTimeout3Minimum = 1 * time.Second
	IdleTimeout3Maximum = 48 * time.Hour

	// "k" range[1, 32767] default value 12. See IEC 60870-5-104, subclass 5.5.
	SendUnAckLimitKMin = 1
	SendUnAckLimitKMax = 32767

	// "w" range[1， 32767] default value 8. See IEC 60870-5-104, subclass 5.5.
	RecvUnAckLimitWMin = 1
	RecvUnAckLimitWMax = 32767
)

// Config defines an IEC 60870-5-104 configuration.
// If not valid configuration is set,the default configuration will be applied.
type Config struct {
	// Maximum timeout for tcp connection establishment
	// "t0" range[1, 255]s, default value 30s.
	ConnectTimeout0 time.Duration

	// I-frames send the upper limit of the number of frames that have not received confirmation.
	// Once this number of frames is reached, transmission will stop.
	// "k" range[1, 32767] default value 12.
	// See IEC 60870-5-104, subclass 5.5.
	SendUnAckLimitK uint16

	// t1 value, is the maximum timeout for frame reception confirmation.
	// The connection will be closed immediately after this time.
	// "t1" range[1, 255]s default value 15s.
	// See IEC 60870-5-104, figure 18.
	SendUnAckTimeout1 time.Duration

	// The receiving end shall issue an acknowledgment after receiving "w" I-frames application protocol data units at the latest.
	//  w value is recommended to not exceed 2/3*k
	// "w" range[1， 32767] default value 8.
	// See IEC 60870-5-104, subclass 5.5.
	RecvUnAckLimitW uint16

	// t2 is the maximum time to send a received confirmation.
	// "t2" range[1, 255]s default value 10s
	// See IEC 60870-5-104, figure 10.
	RecvUnAckTimeout2 time.Duration

	//t3 is the idle time value that triggers "TESTFR" keepalive message.
	// "t3" range[1 second, 48 hours] default value 20 s
	// See IEC 60870-5-104, subclass 5.2.
	IdleTimeout3 time.Duration
}

// ValidConfigServer check if the configuration is valid, if not returns an error.
func (sf *Config) ValidConfigServer() error {
	if sf == nil {
		return errors.New("invalid pointer to server configuration")
	}

	if sf.ConnectTimeout0 < ConnectTimeout0Min || sf.ConnectTimeout0 > ConnectTimeout0Max {
		return errors.New(`ConnectTimeout0 "t0" is not in the correct range [1, 255]s`)
	}

	if sf.SendUnAckLimitK == 0 {
		sf.SendUnAckLimitK = 12
	} else if sf.SendUnAckLimitK < SendUnAckLimitKMin || sf.SendUnAckLimitK > SendUnAckLimitKMax {
		return errors.New(`SendUnAckLimitK "k" not in [1, 32767]`)
	}

	if sf.SendUnAckTimeout1 == 0 {
		sf.SendUnAckTimeout1 = 15 * time.Second
	} else if sf.SendUnAckTimeout1 < SendUnAckTimeout1Min || sf.SendUnAckTimeout1 > SendUnAckTimeout1Max {
		return errors.New(`SendUnAckTimeout1 "t₁" not in [1, 255]s`)
	}

	if sf.RecvUnAckLimitW == 0 {
		sf.RecvUnAckLimitW = 8
	} else if sf.RecvUnAckLimitW < RecvUnAckLimitWMin || sf.RecvUnAckLimitW > RecvUnAckLimitWMax {
		return errors.New(`RecvUnAckLimitW "w" not in [1, 32767]`)
	}

	if sf.RecvUnAckTimeout2 == 0 {
		sf.RecvUnAckTimeout2 = 10 * time.Second
	} else if sf.RecvUnAckTimeout2 < RecvUnAckTimeout2Min || sf.RecvUnAckTimeout2 > RecvUnAckTimeout2Max {
		return errors.New(`RecvUnAckTimeout2 "t₂" not in [1, 255]s`)
	}

	if sf.IdleTimeout3 == 0 {
		sf.IdleTimeout3 = 20 * time.Second
	} else if sf.IdleTimeout3 < IdleTimeout3Minimum || sf.IdleTimeout3 > IdleTimeout3Maximum {
		return errors.New(`IdleTimeout3 "t₃" not in [1 second, 48 hours]`)
	}

	return nil
}

// DefaultConfig will apply the default configuration to the server.
func DefaultConfig() Config {
	return Config{
		ConnectTimeout0:   30 * time.Second,
		SendUnAckLimitK:   12,
		SendUnAckTimeout1: 15 * time.Second,
		RecvUnAckLimitW:   8,
		RecvUnAckTimeout2: 10 * time.Second,
		IdleTimeout3:      20 * time.Second,
	}
}
