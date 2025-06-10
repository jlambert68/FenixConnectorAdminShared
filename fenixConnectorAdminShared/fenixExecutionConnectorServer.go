package fenixConnectorAdminShared

import (
	"fmt"
	uuidGenerator "github.com/google/uuid"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/connectorEngine"
	"github.com/jlambert68/FenixConnectorAdminShared/gRPCServer"
	"github.com/jlambert68/FenixConnectorAdminShared/incomingPubSubMessages"
	"github.com/jlambert68/FenixConnectorAdminShared/messagesToExecutionWorkerServer"
	"github.com/sirupsen/logrus"
	"time"
)

// Used for only process cleanup once
var cleanupProcessed = false

func cleanup(stopAliveToWorkerTickerChannel *chan common_config.StopAliveToWorkerTickerChannelStruct) {

	if cleanupProcessed == false {

		// Stop Ticker used when informing Worker that Connector is alive
		var stopAliveToWorkerTickerChannelMessage common_config.StopAliveToWorkerTickerChannelStruct
		var returnChannel chan bool
		returnChannel = make(chan bool)

		stopAliveToWorkerTickerChannelMessage = common_config.StopAliveToWorkerTickerChannelStruct{
			ReturnChannel: &returnChannel}

		// Send Message twice due to logic in receiver side
		*stopAliveToWorkerTickerChannel <- stopAliveToWorkerTickerChannelMessage
		*stopAliveToWorkerTickerChannel <- stopAliveToWorkerTickerChannelMessage

		// Wait for message has been sent to Worker
		<-returnChannel

		// Inform Worker that Connector is closing down
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.ConnectorIsShuttingDown()

		cleanupProcessed = true

		// Cleanup before close down application
		fenixConnectorAdminSharedObject.logger.WithFields(logrus.Fields{
			"id": "5a76a42c-66af-464a-a62c-f45f7c3fa2d5",
		}).Info("Clean up and shut down servers")

	}
}

func fenixExecutionConnectorMain() {

	// Create Unique Uuid for run time instance used as identification when communication with GuiExecutionServer
	common_config.ApplicationRunTimeUuid = uuidGenerator.New().String()
	fmt.Println("sharedCode.ApplicationRunTimeUuid: " + common_config.ApplicationRunTimeUuid)

	// Set up BackendObject
	fenixConnectorAdminSharedObject = &fenixConnectorAdminSharedObjectStruct{
		logger: nil,
		TestInstructionExecutionEngine: connectorEngine.TestInstructionExecutionEngineStruct{
			MessagesToExecutionWorkerObjectReference: &messagesToExecutionWorkerServer.MessagesToExecutionWorkerObjectStruct{
				//GcpAccessToken: nil,
			},
		},
	}

	connectorEngine.TestInstructionExecutionEngine = connectorEngine.TestInstructionExecutionEngineStruct{
		MessagesToExecutionWorkerObjectReference: &messagesToExecutionWorkerServer.MessagesToExecutionWorkerObjectStruct{
			//GcpAccessToken: nil,
		},
	}

	// Init logger
	//fenixConnectorAdminSharedObject.InitLogger(loggerFileName)
	fenixConnectorAdminSharedObject.logger = common_config.Logger

	// Clean up when leaving. Is placed after logger because shutdown logs information
	// Channel is used for syncing messages: "Connector is Ready for Work" vs "Connector is shutting down"
	var stopSendingAliveToWorkerTickerChannel chan common_config.StopAliveToWorkerTickerChannelStruct
	stopSendingAliveToWorkerTickerChannel = make(chan common_config.StopAliveToWorkerTickerChannelStruct)
	defer cleanup(&stopSendingAliveToWorkerTickerChannel)

	// Initiate CommandChannel
	connectorEngine.ExecutionEngineCommandChannel = make(chan connectorEngine.ChannelCommandStruct)

	// Start ChannelCommand Engine
	fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.CommandChannelReference = &connectorEngine.ExecutionEngineCommandChannel
	fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.InitiateTestInstructionExecutionEngineCommandChannelReader(connectorEngine.ExecutionEngineCommandChannel)

	// Channel for informing that an access token was received
	var connectorIsReadyToReceiveWorkChannel chan bool
	connectorIsReadyToReceiveWorkChannel = make(chan bool)

	// 	Inform Worker that Connector is ready to receive work
	go func() {

		// Send Supported TestInstructions, TesInstructionContainers and Allowed Users to Worker
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers()

		//Send template repository connection parameters to Worker
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendTemplateRepositoryConnectionParameters()

		// Send 'simple' TestData to Worker
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendSimpleTestData()

		// Send SupportTestCaseMetaData and SupportTestSuiteMetaData to worker
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendSupportedMetaData()

		// Should there be any communication out from Connector
		if common_config.TurnOffAllCommunicationWithWorker == false {

			// Wait 5 seconds before informing Worker that Connector is ready for Work
			time.Sleep(5 * time.Second)
			// Inform Worker that Connector is Starting Up
			fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
				ConnectorIsReadyToReceiveWork(&stopSendingAliveToWorkerTickerChannel, &connectorIsReadyToReceiveWorkChannel)
		}

	}()

	// Start up PubSub-receiver, if it should
	if common_config.ShouldPubSubReceiverBeStarted == true && common_config.TurnOffAllCommunicationWithWorker == false {

		if common_config.UsePubSubToReceiveMessagesFromWorker == true {

			if common_config.UseNativeGcpPubSubClientLibrary == true {
				// Use Native GCP PubSub Client Library
				go incomingPubSubMessages.PullPubSubTestInstructionExecutionMessagesGcpClientLib(&connectorIsReadyToReceiveWorkChannel)

			} else {
				// Use REST to call GCP PubSub
				go incomingPubSubMessages.PullPubSubTestInstructionExecutionMessagesGcpRestApi(&connectorIsReadyToReceiveWorkChannel)

			}
		}
	}

	// Initiate and start the Worker gRPC-server
	gRPCServer.FenixConnectorGrpcServicesServerObject = &gRPCServer.FenixConnectorGrpcServicesServerStruct{}
	gRPCServer.FenixConnectorGrpcServicesServerObject.InitGrpcServer()

}
