package common_config

import (
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"time"
)

// ConnectorCallBackFunctionsStruct
// Struct holding the functions that can be used via call-back to Connector
type ConnectorCallBackFunctionsStruct struct {
	// Gets the max time for when the TestInstructionExecution can be seen as "dead"
	GetMaxExpectedFinishedTimeStamp func() (
		maxExpectedFinishedTimeStamp time.Time,
		err error)

	// Sends the TestInstruction to be executed by Connector code unique to every system
	ProcessTestInstructionExecutionRequest func(
		testInstructionExecutionPubSubRequest *fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest) (
		testInstructionExecutionResultMessage *fenixExecutionWorkerGrpcApi.FinalTestInstructionExecutionResultMessage,
		err error)

	// Initiate callers logger
	InitiateLogger func(
		logger *logrus.Logger)
}

// ConnectorFunctionsToDoCallBackOn
// Variable that stores the functions that can be call via call-back
var ConnectorFunctionsToDoCallBackOn *ConnectorCallBackFunctionsStruct
