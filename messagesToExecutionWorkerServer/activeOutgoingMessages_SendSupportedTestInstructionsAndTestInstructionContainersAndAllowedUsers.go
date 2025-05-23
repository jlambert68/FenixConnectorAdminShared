package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/TestInstructionAndTestInstuctionContainerTypes"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/TypeAndStructs"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/shared_code"
	"github.com/sirupsen/logrus"
	"time"
)

// SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers
// Send supported TestInstructions, TestInstructionContainers and Allowed Users to Worker
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers() {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "6c25abd5-9130-4cfc-b26d-88c21852d5ba",
	}).Debug("Incoming 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "cfa9f898-ece2-44b1-aa93-3d9010175143",
	}).Debug("Outgoing 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")

	var err error

	// Do call-back to get all 	// Create supported TestInstructions, TestInstructionContainers and Allowed Users
	var supportedTestInstructionsAndTestInstructionContainersAndAllowedUsers *TestInstructionAndTestInstuctionContainerTypes.
		TestInstructionsAndTestInstructionsContainersStruct
	supportedTestInstructionsAndTestInstructionContainersAndAllowedUsers = common_config.ConnectorFunctionsToDoCallBackOn.
		GenerateSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers()

	// Check if this Connector is the one that sends Supported TestInstructions, TesInstructionContainers and
	// Allowed Users to Worker. If not then just exit
	if common_config.ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers == false {
		return
	}

	// Make override on if a New Baseline should be saved in database for
	// TestInstructions, TestInstructionContainers and Allowed Users if environment variable is set
	if common_config.ForceNewBaseLineForTestInstructionsAndTestInstructionContainers == true {
		supportedTestInstructionsAndTestInstructionContainersAndAllowedUsers.
			ForceNewBaseLineForTestInstructionsAndTestInstructionContainers = true
	}

	// Convert supported TestInstructions, TestInstructionContainers and Allowed Users message into a gRPC-Worker version of the message
	var supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage *fenixExecutionWorkerGrpcApi.
		SupportedTestInstructionsAndTestInstructionContainersAndAllowedUsersMessage
	supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage, err = shared_code.
		GenerateTestInstructionAndTestInstructionContainerAndUserGrpcWorkerMessage(
			common_config.ThisDomainsUuid,
			supportedTestInstructionsAndTestInstructionContainersAndAllowedUsers)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":    "d44f1c6e-bc01-4d2e-a511-f6b04e03515d",
			"error": err,
		}).Fatalln("Problem when generating 'supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage' " +
			"in 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")
	}

	// Convert back supported TestInstructions, TestInstructionContainers and Allowed Users message from a gRPC-Worker version of the message and check correctness of Hashes
	var testInstructionsAndTestInstructionContainersFromGrpcWorkerMessage *TestInstructionAndTestInstuctionContainerTypes.
		TestInstructionsAndTestInstructionsContainersStruct
	testInstructionsAndTestInstructionContainersFromGrpcWorkerMessage, err = shared_code.
		GenerateStandardFromGrpcWorkerMessageForTestInstructionsAndUsers(
			supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":    "4590b988-c944-44d6-80c3-69368bb0046f",
			"error": err,
		}).Fatalln("Problem when Convert back supported TestInstructions, TestInstructionContainers and " +
			"Allowed Users message from a gRPC-Worker version of the message and check correctness of Hashes " +
			"in 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")
	}

	// Verify recreated Hashes from gRPC-Worker-message
	var errorSliceWorker []error
	errorSliceWorker = shared_code.VerifyTestInstructionAndTestInstructionContainerAndUsersMessageHashesAndDomain(
		TypeAndStructs.DomainUUIDType(common_config.ThisDomainsUuid),
		testInstructionsAndTestInstructionContainersFromGrpcWorkerMessage)
	if errorSliceWorker != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":               "2a2cd79b-5694-4562-ad60-21d16e087c53",
			"errorSliceWorker": errorSliceWorker,
		}).Fatalln("Problem when recreated Hashes from gRPC-Worker-message " +
			"in 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")
	}

	// Generate the public key used to verify the signature
	var publicKeyAsBase64String string
	publicKeyAsBase64String, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKey()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "d5a6735d-22ed-4de7-a2c8-231c9109f3cb",
			"err": err,
		}).Fatalln("Couldn't generate Public key from Private key Message")
	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID":                              "0ef03b7c-cd65-4065-b503-78e25bae2dd6",
		"messageHashToSign":               supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage.MessageSignatureData.GetHashToBeSigned(),
		"publicKeyAsBase64String":         publicKeyAsBase64String,
		"signatureToVerifyAsBase64String": supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage.MessageSignatureData.GetSignature(),
	}).Info("Message to be signed, Signature and public key")

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
			"ID":    "b87f3a70-689e-4888-b012-837d5585e72e",
			"error": err,
		}).Fatalln("Problem setting up connection to Fenix Execution Worker for 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")
	}

	// Do gRPC-call
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		common_config.Logger.WithFields(logrus.Fields{
			"ID": "34c7efe8-e1ab-4b6a-a945-59727f730a2e",
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
				"ID":                  "0b2d4659-9078-4c24-93b3-40cd5f213f98",
				"returnMessageString": returnMessageString,
			}).Fatalln("Problem generating GCP access token for 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")
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
			ConnectorPublishSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers(
				ctx,
				supportedTestInstructionsAndTestInstructionContainersAndAllowedUsersGrpcWorkerMessage)

		// Add to counter for how many gRPC-call-attempts to Worker that have been done
		gRPCCallAttemptCounter = gRPCCallAttemptCounter + 1

		// Shouldn't happen
		if err != nil {

			// Only return the error after last attempt
			if gRPCCallAttemptCounter >= numberOfgRPCCallAttempts {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":    "c54f7b47-efa2-44a7-a73b-8907d338b4b6",
					"error": err,
				}).Fatalln("Problem to do gRPC-call to Fenix Execution Worker for 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")

			}

			// Sleep for some time before retrying to connect
			time.Sleep(time.Millisecond * time.Duration(sleepTimeBetweenGrpcCallAttempts[gRPCCallAttemptCounter-1]))

		} else if returnMessage.AckNack == false {
			// Couldn't handle gPRC call
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                        "1adae83c-b5fd-4b27-a3e5-23699d9b5d18",
				"Message from Fenix Worker": returnMessage.Comments,
			}).Fatalln("Problem to do gRPC-call to Worker for 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers'")

		} else {

			common_config.Logger.WithFields(logrus.Fields{
				"ID": "6acf790a-899e-4e94-9858-1760d12e2fc7",
			}).Debug("Success in doing gRPC-call to Worker for 'SendSupportedTestInstructionsAndTestInstructionContainersAndAllowedUsers")

			return

		}

	}
}
