// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"time"

	"gitlab.com/circutor-library/go-iecp5/asdu"
)

// ServerHandlerInterface is the interface of server handler
type ServerHandlerInterface interface {
	InterrogationHandler(con asdu.Connect, asdu *asdu.ASDU, qualifierInt asdu.QualifierOfInterrogation) error
	CounterInterrogationHandler(con asdu.Connect, asdu *asdu.ASDU, counterInt asdu.QualifierCountCall) error
	ReadHandler(con asdu.Connect, asdu *asdu.ASDU, readInfo asdu.InfoObjAddr) error
	ClockSyncHandler(con asdu.Connect, asdu *asdu.ASDU, time time.Time) error
	ResetProcessHandler(con asdu.Connect, asdu *asdu.ASDU, resetQuaifier asdu.QualifierOfResetProcessCmd) error
	DelayAcquisitionHandler(con asdu.Connect, asdu *asdu.ASDU, delay uint16) error
	ASDUHandler(con asdu.Connect, asdu *asdu.ASDU) error
	DoubleCommandHandler(con asdu.Connect, asdu *asdu.ASDU, doubleCom asdu.DoubleCommandInfo) error
	SingleCommandHandler(con asdu.Connect, asdu *asdu.ASDU, singleCom asdu.SingleCommandInfo) error
	SetPointCommandNormalHandler(con asdu.Connect, asdu *asdu.ASDU, setNormal asdu.SetpointCommandNormalInfo) error
	SetPointCommandScaledHandler(con asdu.Connect, asdu *asdu.ASDU, setScaled asdu.SetpointCommandScaledInfo) error
}

// ServerQueueManagerInterface is the manager of connection queues
type ServerQueueManagerInterface interface {
	NewQueue(remoteAddr string) ServerQueueInterface
	DeleteQueue(remoteAddr string)
}

// ServerQueueInterface is the interface of connection queue
type ServerQueueInterface interface {
	Enqueue(con asdu.Connect, frame asdu.ASDU) error
	Dequeue() (asdu.Connect, asdu.ASDU, error)
}

// ClientHandlerInterface  is the interface of client handler
type ClientHandlerInterface interface {
	InterrogationHandler(con asdu.Connect, asdu *asdu.ASDU) error
	CounterInterrogationHandler(con asdu.Connect, asdu *asdu.ASDU) error
	ReadHandler(con asdu.Connect, asdu *asdu.ASDU) error
	TestCommandHandler(con asdu.Connect, asdu *asdu.ASDU) error
	ClockSyncHandler(con asdu.Connect, asdu *asdu.ASDU) error
	ResetProcessHandler(con asdu.Connect, asdu *asdu.ASDU) error
	DelayAcquisitionHandler(con asdu.Connect, asdu *asdu.ASDU) error
	ASDUHandler(con asdu.Connect, asdu *asdu.ASDU) error
}
