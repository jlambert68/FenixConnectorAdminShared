package common_config

import (
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/jlambert68/FenixScriptEngine/testDataEngine"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/TestInstructionAndTestInstuctionContainerTypes"
	"github.com/sirupsen/logrus"
	"time"
)

// ConnectorCallBackFunctionsStruct
// Struct holding the functions that can be used via call-back to Connector
type ConnectorCallBackFunctionsStruct struct {
	// Gets the max time for when the TestInstructionExecution can be seen as "dead"
	GetMaxExpectedFinishedTimeStamp func(
		testInstructionExecutionPubSubRequest *fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest) (
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

	// Generates the 'SupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers' that will be sent via gRPC to Worker
	GenerateSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers func() (
		supportedTestInstructionsAndTestInstructionContainersAndAllowedUsers *TestInstructionAndTestInstuctionContainerTypes.
			TestInstructionsAndTestInstructionsContainersStruct,
	)

	// Sends this Github url parameters to where Templates are stored
	GenerateTemplateRepositoryConnectionParameters func() (
		templateRepositoryConnectionParameters *RepositoryTemplatePathStruct)

	// Sends "Simple" TestData towards TestCaseBuilderServer
	GenerateSimpleTestData func() (
		simpleTestData []*testDataEngine.TestDataFromSimpleTestDataAreaStruct)

	// Sends "Simple" TestData towards TestCaseBuilderServer
	GenerateSupportedMetaData func() (
		supportedMetaData *[]byte)
}

// ConnectorFunctionsToDoCallBackOn
// Variable that stores the functions that can be call via call-back
var ConnectorFunctionsToDoCallBackOn *ConnectorCallBackFunctionsStruct
