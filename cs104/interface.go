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
	SingleCommandHandler(con asdu.Connect, asdu *asdu.ASDU, singleCom asdu.SingleCommandInfo, cmdInfo asdu.InfoObjAddr) asdu.CauseOfTransmission
	DoubleCommandHandler(con asdu.Connect, asdu *asdu.ASDU, doubleCom asdu.DoubleCommandInfo, cmdInfo asdu.InfoObjAddr) asdu.CauseOfTransmission
	StepPositionCommandHandler(con asdu.Connect, asdu *asdu.ASDU, setScaled asdu.StepCommandInfo, cmdInfo asdu.InfoObjAddr) asdu.CauseOfTransmission
	SetPointCommandFloatHandler(con asdu.Connect, asdu *asdu.ASDU, setFloat asdu.SetpointCommandFloatInfo, cmdInfo asdu.InfoObjAddr) asdu.CauseOfTransmission
	InterrogationHandler(con asdu.Connect, asdu *asdu.ASDU, qualifierInt asdu.QualifierOfInterrogation) (cot asdu.CauseOfTransmission, ASDUAddr int)
	ReadHandler(con asdu.Connect, asdu *asdu.ASDU, readInfo asdu.InfoObjAddr) error
	ResetProcessHandler(con asdu.Connect, asdu *asdu.ASDU, resetQualifier asdu.QualifierOfResetProcessCmd) asdu.CauseOfTransmission
	ClockSyncHandler(con asdu.Connect, asdu *asdu.ASDU, time time.Time) asdu.CauseOfTransmission
	ASDUHandler(con asdu.Connect, asdu *asdu.ASDU) error
}

// ServerQueueInterface is the interface of connection queue
type ServerQueueInterface interface {
	Enqueue(frame asdu.ASDU) error
	Dequeue() (asdu.ASDU, error)
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
