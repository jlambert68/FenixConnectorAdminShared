package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	"github.com/jlambert68/FenixConnectorAdminShared/supportedSubInstructions"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	fenixSyncShared "github.com/jlambert68/FenixSyncShared"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/shared_code"
	"github.com/sirupsen/logrus"
	"time"
)

// SendSupportedSubInstructions
// Send the supported SubInstructions to Worker
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendSupportedSubInstructions() {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "bf451cf7-00a4-4fe6-9603-5f9f6a892106",
	}).Debug("Incoming 'SendSupportedSubInstructions'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "3e145b6a-3963-4426-b17d-79572ab82b7a",
	}).Debug("Outgoing 'SendSupportedSubInstructions'")

	var err error

	// Do call-back to get all, TestCase and TestSuite, MetaData that should be sent
	var supportedSubInstructionsAsByteSlice *[]byte
	var supportedSubInstructionsPerTestInstructionAsByteSlice *[][]byte
	supportedSubInstructionsAsByteSlice = common_config.ConnectorFunctionsToDoCallBackOn.GenerateSupportedSubInstructions()
	supportedSubInstructionsPerTestInstructionAsByteSlice = common_config.ConnectorFunctionsToDoCallBackOn.GenerateSupportedSubInstructionsPerTestInstruction()

	// Convert the '[]byte' into a 'string'
	var supportedSubInstructionsAsString string
	var supportedSubInstructionsPerTestInstructionAsStringSlice []string
	supportedSubInstructionsAsString = string(*supportedSubInstructionsAsByteSlice)
	for _, tempSupportedSubInstructionsPerTestInstructionAsByteSlice := range *supportedSubInstructionsPerTestInstructionAsByteSlice {
		supportedSubInstructionsPerTestInstructionAsStringSlice = append(
			supportedSubInstructionsPerTestInstructionAsStringSlice, string(tempSupportedSubInstructionsPerTestInstructionAsByteSlice))
	}

	// Verify json towards json-schema - SupportedSubInstructions
	err = supportedSubInstructions.ValidateSupportedSubInstructions(supportedSubInstructionsAsByteSlice)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "bf8e6343-b93d-455f-9849-f13674a13567",
			"err": err,
		}).Fatalln("Error when validation 'SendSupportedSubInstructions'")
	}

	// Verify json towards json-schema - TestSuiteMetaData
	err = supportedSubInstructions.ValidateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema(supportedSubInstructionsPerTestInstructionAsByteSlice)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "13124d05-dc67-407a-a9cd-40f0fa41758d",
			"err": err,
		}).Fatalln("Couldn't validate json-message using json-schema for 'SendSupportedSubInstructionsPerTestInstruction'")
	}

	// Calculate the hash for SupportedSubInstructions
	var supportedSubInstructionsHash string
	var subInstructionsToBeHashed []string
	subInstructionsToBeHashed = append(subInstructionsToBeHashed, supportedSubInstructionsAsString)
	subInstructionsToBeHashed = append(subInstructionsToBeHashed, supportedSubInstructionsPerTestInstructionAsStringSlice...)
	supportedSubInstructionsHash = fenixSyncShared.HashValues(subInstructionsToBeHashed, true)

	// Convert into gRPC-message

	// Add Domain-information
	var tempClientSystemIdentificationMessage *fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage
	tempClientSystemIdentificationMessage = &fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage{
		DomainUuid:          common_config.ThisDomainsUuid,
		ExecutionDomainUuid: common_config.ThisExecutionDomainUuid,
		ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(
			common_config.GetHighestExecutionWorkerProtoFileVersion()),
	}

	// Create and sign message
	var messageHashToSign string
	messageHashToSign = supportedSubInstructionsHash

	// Sign the message
	var signatureToVerifyAsBase64String string
	signatureToVerifyAsBase64String, err = shared_code.SignMessageUsingSchnorrSignature(messageHashToSign)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "9fda8c05-1dcc-4a5b-8343-6d6c74edb5cf",
			"err": err,
		}).Fatalln("Couldn't sign Message")
	}

	// Generate the public key used to verify the signature
	var publicKeyAsBase64String string
	publicKeyAsBase64String, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKey()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "3e41c25d-4905-41b9-a480-796b56956e24",
			"err": err,
		}).Fatalln("Couldn't generate Public key from Private key Message")
	}
	// Verify Signature
	err = shared_code.VerifySchnorrSignature(messageHashToSign, publicKeyAsBase64String, signatureToVerifyAsBase64String)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "4ed3b6e8-422d-4012-aed4-5d8c8041a548",
			"err": err,
		}).Fatalln("Couldn't verify the Signature")
	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID":                              "dcbe6e25-c794-45be-b71c-bfdd421a5a40",
		"messageHashToSign":               messageHashToSign,
		"publicKeyAsBase64String":         publicKeyAsBase64String,
		"signatureToVerifyAsBase64String": signatureToVerifyAsBase64String,
	}).Info("Message to be signed, Signature and public key")

	var messageSignatureData *fenixExecutionWorkerGrpcApi.MessageSignatureDataMessage
	messageSignatureData = &fenixExecutionWorkerGrpcApi.MessageSignatureDataMessage{
		HashToBeSigned: messageHashToSign,
		Signature:      signatureToVerifyAsBase64String,
	}

	// Create the full gRPC-message
	var supportedSubInstructionsAsGrpc *fenixExecutionWorkerGrpcApi.SupportedSubInstructions
	supportedSubInstructionsAsGrpc = &fenixExecutionWorkerGrpcApi.SupportedSubInstructions{
		ClientSystemIdentification:                       tempClientSystemIdentificationMessage,
		SupportedSubInstructionsAsJson:                   supportedSubInstructionsAsString,
		SupportedSubInstructionsPerTestInstructionAsJson: supportedSubInstructionsPerTestInstructionAsStringSlice,
		MessageSignatureData:                             messageSignatureData,
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
			"ID":    "c28e37c3-4c09-4e9a-9a09-b1874c0b56ff",
			"error": err,
		}).Fatalln("Problem setting up connection to Fenix Execution Worker for 'SendSupportedSubInstructions'")
	}

	// Do gRPC-call
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		common_config.Logger.WithFields(logrus.Fields{
			"ID": "ea109a7a-b640-4905-a5f4-9f03cebf3b95",
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
				"ID":                  "3d9cb846-3b12-4124-9c9d-26ce4b8545de",
				"returnMessageString": returnMessageString,
			}).Fatalln("Problem generating GCP access token for 'SendSupportedSubInstructions'")
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
			ConnectorPublishSupportedSubInstructions(
				ctx,
				supportedSubInstructionsAsGrpc)

		// Add to counter for how many gRPC-call-attempts to Worker that have been done
		gRPCCallAttemptCounter = gRPCCallAttemptCounter + 1

		// Shouldn't happen
		if err != nil {

			// Only return the error after last attempt
			if gRPCCallAttemptCounter >= numberOfgRPCCallAttempts {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":    "27f07a79-3ac5-4466-a935-78ba8ffce803",
					"error": err,
				}).Fatalln("Problem to do gRPC-call to Fenix Execution Worker for 'SendSupportedSubInstructions'")

			}

			// Sleep for some time before retrying to connect
			time.Sleep(time.Millisecond * time.Duration(sleepTimeBetweenGrpcCallAttempts[gRPCCallAttemptCounter-1]))

		} else if returnMessage.AckNack == false {
			// Couldn't handle gPRC call
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                        "f6911973-74b2-470a-b914-88808d258cdc",
				"Message from Fenix Worker": returnMessage.Comments,
			}).Fatalln("Problem to do gRPC-call to Worker for 'SendSupportedSubInstructions'")

		} else {

			common_config.Logger.WithFields(logrus.Fields{
				"ID": "edec2763-8f5e-422c-a8b2-bb0d00056e84",
			}).Debug("Success in doing gRPC-call to Worker for 'SendSupportedSubInstructions")

			return

		}

	}
}
