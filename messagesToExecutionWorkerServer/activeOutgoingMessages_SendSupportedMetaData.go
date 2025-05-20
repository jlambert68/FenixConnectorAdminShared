package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	"github.com/jlambert68/FenixConnectorAdminShared/supportedMetaData"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	fenixSyncShared "github.com/jlambert68/FenixSyncShared"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/shared_code"
	"github.com/sirupsen/logrus"
	"time"
)

// SendSupportedMetaData
// Send the supported MetaData to Worker
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendSupportedMetaData() {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "e6383b8e-c0e8-4eeb-b0ba-7bac54e65119",
	}).Debug("Incoming 'SendSupportedMetaData'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "87971f33-d484-4aec-8561-ba72afbdabbe",
	}).Debug("Outgoing 'SendSupportedMetaData'")

	var err error

	// Do call-back to get all 	MetaData that should be sent
	var supportedMetaDataAsByteSlice *[]byte
	supportedMetaDataAsByteSlice = common_config.ConnectorFunctionsToDoCallBackOn.GenerateSupportedMetaData()

	// Convert the '[]byte' into a 'string'
	var supportedMetaDataAsString string
	supportedMetaDataAsString = string(*supportedMetaDataAsByteSlice)

	// Verify json towards json-schema
	err = supportedMetaData.ValidateSupportedMetaDataJsonTowardsJsonSchema(&supportedMetaDataAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "b9427f39-8fbf-42d4-bd39-aece016e8369",
			"err": err,
		}).Fatalln("Couldn't sign Message")
	}

	// Calculate the hash for SupportedMetaData
	var supportedMetaDataHash string
	supportedMetaDataHash = fenixSyncShared.HashSingleValue(supportedMetaDataAsString)

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
	messageHashToSign = supportedMetaDataHash

	// Sign the message
	var signatureToVerifyAsBase64String string
	signatureToVerifyAsBase64String, err = shared_code.SignMessageUsingSchnorrSignature(messageHashToSign)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "72b84919-3966-4e08-b475-0e8bedb34856",
			"err": err,
		}).Fatalln("Couldn't sign Message")
	}

	// Generate the public key used to verify the signature
	var publicKeyAsBase64String string
	publicKeyAsBase64String, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKey()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "b1dccf16-a375-47a1-85d5-ab051b629b95",
			"err": err,
		}).Fatalln("Couldn't generate Public key from Private key Message")
	}
	// Verify Signature
	err = shared_code.VerifySchnorrSignature(messageHashToSign, publicKeyAsBase64String, signatureToVerifyAsBase64String)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "625bd417-a395-49e0-aa60-15238cd399c8",
			"err": err,
		}).Fatalln("Couldn't verify the Signature")
	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID":                              "cf51ff69-0d3a-4b8b-aebb-8947f5453fa7",
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
	var supportedMetaDataAsGrpc *fenixExecutionWorkerGrpcApi.SupportedTestCaseMetaData
	supportedMetaDataAsGrpc = &fenixExecutionWorkerGrpcApi.SupportedTestCaseMetaData{
		ClientSystemIdentification: tempClientSystemIdentificationMessage,
		AllTemplateRepositories:    allTemplateRepositories,
		MessageSignatureData:       messageSignatureData,
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
		}).Fatalln("Problem setting up connection to Fenix Execution Worker for 'SendSupportedMetaData'")
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
			}).Fatalln("Problem generating GCP access token for 'SendSupportedMetaData'")
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
			ConnectorPublishSupportedMetaData(
				ctx,
				supportedMetaDataAsGrpc)

		// Add to counter for how many gRPC-call-attempts to Worker that have been done
		gRPCCallAttemptCounter = gRPCCallAttemptCounter + 1

		// Shouldn't happen
		if err != nil {

			// Only return the error after last attempt
			if gRPCCallAttemptCounter >= numberOfgRPCCallAttempts {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":    "bdbd1322-abc7-46fd-aa21-65701d640914",
					"error": err,
				}).Fatalln("Problem to do gRPC-call to Fenix Execution Worker for 'SendSupportedMetaData'")

			}

			// Sleep for some time before retrying to connect
			time.Sleep(time.Millisecond * time.Duration(sleepTimeBetweenGrpcCallAttempts[gRPCCallAttemptCounter-1]))

		} else if returnMessage.AckNack == false {
			// Couldn't handle gPRC call
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                        "f6911973-74b2-470a-b914-88808d258cdc",
				"Message from Fenix Worker": returnMessage.Comments,
			}).Fatalln("Problem to do gRPC-call to Worker for 'SendSupportedMetaData'")

		} else {

			common_config.Logger.WithFields(logrus.Fields{
				"ID": "edec2763-8f5e-422c-a8b2-bb0d00056e84",
			}).Debug("Success in doing gRPC-call to Worker for 'SendSupportedMetaData")

			return

		}

	}
}
