// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs101

// FT1.2 frame format
const (
	StartVarFrame byte = 0x68 // Variable length frame start characters
	StartFixFrame byte = 0x10 // Fixed length frame startup characters
	EndFrame      byte = 0x16
)

// Control field definition
const (

	// Starting station to slave station specific
	FCV = 1 << 4 // Frame Count Valid Bits
	FCB = 1 << 5 // Frame Count Bit
	// Slave-to-start station specific
	DFC     = 1 << 4 // Data flow control bits
	ACD_RES = 1 << 5 // Required access bit, unbalanced ACD, balanced hold
	// Startup message bit: // PRM = 0, slave transmits message to startup station
	// PRM = 0, slave to initiator; // PRM = 1, slave to initiator; // PRM = 1, slave to initiator.
	// PRM = 1, transmit message from slave to initiator; // PRM = 1, transmit message from initiator to slave;
	// PRM = 1, transmit message from initiator to slave
	RPM     = 1 << 6
	RES_DIR = 1 << 7 // Unbalanced hold, balanced for direction
)

const (
	// Function code for the control field in the message transmitted from the initiating station to the slave station (PRM = 1)
	FccResetRemoteLink                 = iota // Reset remote link
	FccResetUserProcess                       // Reset the user process
	FccBalanceTestLink                        // Link test function
	FccUserDataWithConfirmed                  // User data, to be confirmed
	FccUserDataWithUnconfirmed                // User data, no confirmation required
	_                                         // Reservations
	_                                         // Defined by agreement between manufacturer and user
	_                                         // Defined by agreement between manufacturer and user
	FccUnbalanceWithRequestBitResponse        // Respond with a request to access the bit
	FccLinkStatus                             // Request link status
	FccUnbalanceLevel1UserData                // Request Level 1 user data
	FccUnbalanceLevel2UserData                // Request Level 2 user data
	// 12-13: Alternative
	// 14-15: Definition by agreement between manufacturer and user
)

const (
	// Function code for the control field in the message transmitted from the slave to the initiating station (PRM = 0)
	FcsConfirmed                 = iota // Endorsement: Affirmation of endorsement
	FcsNConfirmed                       // Negative acknowledgement: No message received, link is busy.
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Defined by agreement between manufacturer and user
	_                                   // Defined by agreement between manufacturer and user
	FcsUnbalanceResponse                // User data
	FcsUnbalanceNegativeResponse        // Negative Recognition: Nothing to call for
	_                                   // Reserved
	FcsStatus                           // Link status or request for access
	// 12: Alternative
	// 13: Definition negotiated between manufacturer and user
	// 14: Link service not working
	// 15: Link service not completed
)

// Ft12 ...
type Ft12 struct {
	Start        byte
	ApduFiledLen byte
	Ctrl         byte
	Address      uint16
	Checksum     byte
	End          byte
}
