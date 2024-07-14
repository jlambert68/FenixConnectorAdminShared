package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"time"
)

// SendSimpleTestData
// Send TestData based on 'simple' TestData-files is to be sent to Worker
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendSimpleTestData() {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "32c4646c-6051-4f67-96c6-1f440a4d28e3",
	}).Debug("Incoming 'SendSimpleTestData'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "e45697a1-8e17-4be2-b645-7b0d96b5761c",
	}).Debug("Outgoing 'SendSimpleTestData'")

	var err error

	// Do call-back to get all 	// Create supported TestInstructions, TestInstructionContainers and Allowed Users
	var simpleTestData []*fenixExecutionWorkerGrpcApi.TestDataFromOneSimpleTestDataAreaFileMessage
	simpleTestData = common_config.ConnectorFunctionsToDoCallBackOn.GenerateSimpleTestData()

	// If there are no "Simple" TestData then just exist
	if simpleTestData == nil {
		return
	}

	// Add Domain-information
	var tempClientSystemIdentificationMessage *fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage
	tempClientSystemIdentificationMessage = &fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage{
		DomainUuid:          common_config.ThisDomainsUuid,
		ExecutionDomainUuid: common_config.ThisExecutionDomainUuid,
		ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(
			common_config.GetHighestExecutionWorkerProtoFileVersion()),
	}

	// Create the full gRPC-message
	var testDataFromSimpleTestDataAreaFileMessageAsGrpc *fenixExecutionWorkerGrpcApi.TestDataFromSimpleTestDataAreaFileMessage
	testDataFromSimpleTestDataAreaFileMessageAsGrpc = &fenixExecutionWorkerGrpcApi.TestDataFromSimpleTestDataAreaFileMessage{
		ClientSystemIdentification:          tempClientSystemIdentificationMessage,
		TestDataFromSimpleTestDataAreaFiles: simpleTestData,
	}

	// Check if this Connector is the one that sends Supported TestInstructions, TesInstructionContainers,
	// Allowed User and TemplateRepositoryConnectionParameters to Worker. If not then just exit
	if common_config.ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers == false {
		return
	}

	// When there should be no traffic towards Worker then just return
	if common_config.TurnOffAllCommunicationWithWorker == true {
		return
	}

	var ctx context.Context
	var returnMessageAckNack bool
	var returnMessageString string

	ctx = context.Background()

	// Set up connection to Server
	ctx, err = toExecutionWorkerObject.SetConnectionToFenixExecutionWorkerServer(ctx)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":    "305b08c1-7c08-4f0f-9a02-ffc802e28553",
			"error": err,
		}).Fatalln("Problem setting up connection to Fenix Execution Worker for 'SendSimpleTestData'")
	}

	// Do gRPC-call
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		common_config.Logger.WithFields(logrus.Fields{
			"ID": "38880fbe-4780-4d69-9c67-e769648d74d2",
		}).Debug("Running Defer Cancel function")
		cancel()
	}()

	// Only add access token when run on GCP
	if common_config.ExecutionLocationForFenixExecutionWorkerServer == common_config.GCP &&
		common_config.GCPAuthentication == true {

		// Add Access token
		ctx, returnMessageAckNack, returnMessageString = gcp.Gcp.GenerateGCPAccessToken(ctx, gcp.GenerateTokenForGrpcTowardsExecutionWorker)
		if returnMessageAckNack == false {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                  "5a4d988f-d32d-4547-9a21-76aca8ea fa60",
				"returnMessageString": returnMessageString,
			}).Fatalln("Problem generating GCP access token for 'SendSimpleTestData'")
		}

	}

	// slice with sleep time, in milliseconds, between each attempt to do gRPC-call to Worker
	var sleepTimeBetweenGrpcCallAttempts []int
	sleepTimeBetweenGrpcCallAttempts = []int{100, 200, 300, 300, 500, 500, 1000, 1000, 1000, 1000} // Total: 5.9 seconds

	// Do multiple attempts to do gRPC-call to Execution Worker, when it fails
	var numberOfgRPCCallAttempts int
	var gRPCCallAttemptCounter int
	numberOfgRPCCallAttempts = len(sleepTimeBetweenGrpcCallAttempts)
	gRPCCallAttemptCounter = 0

	// Creates a new temporary client only to be used for this call
	//var tempFenixExecutionWorkerConnectorGrpcServicesClient fenixExecutionWorkerGrpcApi.FenixExecutionWorkerConnectorGrpcServicesClient
	//tempFenixExecutionWorkerConnectorGrpcServicesClient = fenixExecutionWorkerGrpcApi.
	//	NewFenixExecutionWorkerConnectorGrpcServicesClient(remoteFenixExecutionWorkerServerConnection)

	for {

		returnMessage, err := fenixExecutionWorkerGrpcClient.
			ConnectorPublishTestDataFromSimpleTestDataAreaFile(
				ctx,
				testDataFromSimpleTestDataAreaFileMessageAsGrpc)

		// Add to counter for how many gRPC-call-attempts to Worker that have been done
		gRPCCallAttemptCounter = gRPCCallAttemptCounter + 1

		// Shouldn't happen
		if err != nil {

			// Only return the error after last attempt
			if gRPCCallAttemptCounter >= numberOfgRPCCallAttempts {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":    "1be898f3-7ba7-4c98-b17b-9d4d2ed06e5c",
					"error": err,
				}).Fatalln("Problem to do gRPC-call to Fenix Execution Worker for 'SendSimpleTestData'")

			}

			// Sleep for some time before retrying to connect
			time.Sleep(time.Millisecond * time.Duration(sleepTimeBetweenGrpcCallAttempts[gRPCCallAttemptCounter-1]))

		} else if returnMessage.AckNack == false {
			// Couldn't handle gPRC call
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                        "4a7da86b-8793-40b2-bb23-b681a2677d11",
				"Message from Fenix Worker": returnMessage.Comments,
			}).Fatalln("Problem to do gRPC-call to Worker for 'SendSimpleTestData'")

		} else {

			common_config.Logger.WithFields(logrus.Fields{
				"ID": "71ec91eb-263a-4a6a-a102-b4b90cb6a7d6",
			}).Debug("Success in doing gRPC-call to Worker for 'SendSimpleTestData")

			return

		}

	}
}
