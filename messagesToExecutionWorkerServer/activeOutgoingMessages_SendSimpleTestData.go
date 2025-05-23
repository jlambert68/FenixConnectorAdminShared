package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/jlambert68/FenixScriptEngine/testDataEngine"
	"github.com/jlambert68/FenixSyncShared"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/shared_code"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
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

	// Do call-back to get all 'Simple' TestData

	var simpleTestData []*testDataEngine.TestDataFromSimpleTestDataAreaStruct
	simpleTestData = common_config.ConnectorFunctionsToDoCallBackOn.GenerateSimpleTestData()

	// If there are no "Simple" TestData then just exist
	if simpleTestData == nil {
		return
	}

	// Generate gRPC-message for Simple TestData-message
	var simpleTestDataAsGrpcMessage []*fenixExecutionWorkerGrpcApi.TestDataFromOneSimpleTestDataAreaFileMessage
	for _, tempTestDataFromSimpleTestDataArea := range simpleTestData {

		// Generate Headers for gRPC-message
		var headersForTestDataFromOneSimpleTestDataAreaFileForGrpc []*fenixExecutionWorkerGrpcApi.
			HeaderForTestDataFromOneSimpleTestDataAreaFileMessage
		for _, header := range tempTestDataFromSimpleTestDataArea.Headers {
			var headerForTestDataFromOneSimpleTestDataAreaFileForGrpc *fenixExecutionWorkerGrpcApi.
				HeaderForTestDataFromOneSimpleTestDataAreaFileMessage
			headerForTestDataFromOneSimpleTestDataAreaFileForGrpc = &fenixExecutionWorkerGrpcApi.
				HeaderForTestDataFromOneSimpleTestDataAreaFileMessage{
				ShouldHeaderActAsFilter: header.ShouldHeaderActAsFilter,
				HeaderName:              header.HeaderName,
				HeaderUiName:            header.HeaderName,
			}

			headersForTestDataFromOneSimpleTestDataAreaFileForGrpc = append(
				headersForTestDataFromOneSimpleTestDataAreaFileForGrpc,
				headerForTestDataFromOneSimpleTestDataAreaFileForGrpc)
		}

		// Generate the TestData-rows for gRPC-message
		var simpleTestDataRowMessageAsGrpc []*fenixExecutionWorkerGrpcApi.SimpleTestDataRowMessage
		var testDataValuesToBeHashed []string
		for _, tempTestDataRow := range tempTestDataFromSimpleTestDataArea.TestDataRows {

			// Convert one row of data into gRPC-version
			var tempTestDataRowAsGrpc *fenixExecutionWorkerGrpcApi.SimpleTestDataRowMessage
			tempTestDataRowAsGrpc = &fenixExecutionWorkerGrpcApi.SimpleTestDataRowMessage{TestDataValue: tempTestDataRow}

			// Add to slice with all TestData to be hashed
			testDataValuesToBeHashed = append(testDataValuesToBeHashed, tempTestDataRow...)

			// Add row to slice of rows
			simpleTestDataRowMessageAsGrpc = append(simpleTestDataRowMessageAsGrpc, tempTestDataRowAsGrpc)
		}

		// Generate ImportantDataInFileSha256Hash
		var importantDataInFileSha256Hash string
		var valuesToHash []string
		valuesToHash = []string{
			tempTestDataFromSimpleTestDataArea.TestDataDomainUuid,
			tempTestDataFromSimpleTestDataArea.TestDataAreaUuid,
		}
		valuesToHash = append(valuesToHash, testDataValuesToBeHashed...)
		importantDataInFileSha256Hash = fenixSyncShared.HashValues(valuesToHash, true)

		// Create the full gRPC-message
		var oneSimpleTestDataAsGrpcMessage *fenixExecutionWorkerGrpcApi.TestDataFromOneSimpleTestDataAreaFileMessage
		oneSimpleTestDataAsGrpcMessage = &fenixExecutionWorkerGrpcApi.TestDataFromOneSimpleTestDataAreaFileMessage{
			TestDataDomainUuid:         tempTestDataFromSimpleTestDataArea.TestDataDomainUuid,
			TestDataDomainName:         tempTestDataFromSimpleTestDataArea.TestDataDomainName,
			TestDataDomainTemplateName: tempTestDataFromSimpleTestDataArea.TestDataDomainTemplateName,
			TestDataAreaUuid:           tempTestDataFromSimpleTestDataArea.TestDataAreaUuid,
			TestDataAreaName:           tempTestDataFromSimpleTestDataArea.TestDataAreaName,
			HeadersForTestDataFromOneSimpleTestDataAreaFile: headersForTestDataFromOneSimpleTestDataAreaFileForGrpc,
			SimpleTestDataRows:            simpleTestDataRowMessageAsGrpc,
			TestDataFileSha256Hash:        tempTestDataFromSimpleTestDataArea.TestDataFileSha256Hash,
			ImportantDataInFileSha256Hash: importantDataInFileSha256Hash,
		}

		// Append to slice
		simpleTestDataAsGrpcMessage = append(simpleTestDataAsGrpcMessage, oneSimpleTestDataAsGrpcMessage)
	}

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
	var hashesToHash []string

	// Loop all TestData and convert into json
	for _, tempTestData := range simpleTestDataAsGrpcMessage {
		var tempTestDataAsJson string
		tempTestDataAsJson = protojson.Format(tempTestData)

		// Remove spaces in json
		tempTestDataAsJson = strings.ReplaceAll(tempTestDataAsJson, " ", "")

		// Append to slice to be hashed
		hashesToHash = append(hashesToHash, tempTestDataAsJson)

	}

	// Create a hash of the slice
	messageHashToSign = fenixSyncShared.HashValues(hashesToHash, true)

	// Sign the message
	var signatureToVerifyAsBase64String string
	signatureToVerifyAsBase64String, err = shared_code.SignMessageUsingSchnorrSignature(messageHashToSign)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "acd02d74-785f-497e-80fb-98b539e755b1",
			"err": err,
		}).Fatalln("Couldn't sign Message")
	}

	// Generate the public key used to verify the signature
	var publicKeyAsBase64String string
	publicKeyAsBase64String, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKey()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "d2ea02e5-b48a-45a7-8fc2-5bd65f298a45",
			"err": err,
		}).Fatalln("Couldn't generate Public key from Private key Message")
	}
	// Verify Signature
	err = shared_code.VerifySchnorrSignature(messageHashToSign, publicKeyAsBase64String, signatureToVerifyAsBase64String)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "7c574dd4-747e-4efa-a703-6d66837c95e1",
			"err": err,
		}).Fatalln("Couldn't verify the Signature")
	}

	common_config.Logger.WithFields(logrus.Fields{
		"ID":                              "f6ab475a-e05d-4da8-9549-df20dade9ce3",
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
	var testDataFromSimpleTestDataAreaFileMessageAsGrpc *fenixExecutionWorkerGrpcApi.TestDataFromSimpleTestDataAreaFileMessage
	testDataFromSimpleTestDataAreaFileMessageAsGrpc = &fenixExecutionWorkerGrpcApi.TestDataFromSimpleTestDataAreaFileMessage{
		ClientSystemIdentification:          tempClientSystemIdentificationMessage,
		TestDataFromSimpleTestDataAreaFiles: simpleTestDataAsGrpcMessage,
		MessageSignatureData:                messageSignatureData,
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
