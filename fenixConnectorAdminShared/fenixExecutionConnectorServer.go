package fenixConnectorAdminShared

import (
	"fmt"
	uuidGenerator "github.com/google/uuid"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/connectorEngine"
	"github.com/jlambert68/FenixConnectorAdminShared/incomingPubSubMessages"
	"github.com/jlambert68/FenixConnectorAdminShared/messagesToExecutionWorkerServer"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
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
		fenixConnectorAdminSharedObject.logger.WithFields(logrus.Fields{}).Info("Clean up and shut down servers")

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
	var accessTokenWasReceivedChannel chan bool
	accessTokenWasReceivedChannel = make(chan bool)

	// 	Inform Worker that Connector is ready to receive work
	go func() {

		// Send Supported TestInstructions, TesInstructionContainers and Allowed Users to Worker
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers()

		// Wait 5 seconds before informing Worker that Connector is ready for Work
		time.Sleep(5 * time.Second)
		// Inform Worker that Connector is Starting Up
		fenixConnectorAdminSharedObject.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			ConnectorIsReadyToReceiveWork(&stopSendingAliveToWorkerTickerChannel, &accessTokenWasReceivedChannel)

	}()

	// Start up PubSub-receiver
	if common_config.UsePubSubToReceiveMessagesFromWorker == true {

		if common_config.UseNativeGcpPubSubClientLibrary == true {
			// Use Native GCP PubSub Client Library
			go incomingPubSubMessages.PullPubSubTestInstructionExecutionMessagesGcpClientLib(&accessTokenWasReceivedChannel)

		} else {
			// Use REST to call GCP PubSub
			go incomingPubSubMessages.PullPubSubTestInstructionExecutionMessagesGcpRestApi(&accessTokenWasReceivedChannel)

		}
	}

	// Wait for 'ctrl c' to exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(&stopSendingAliveToWorkerTickerChannel)
		os.Exit(0)
	}()

}
