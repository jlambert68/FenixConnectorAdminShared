package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"time"
)

// SendTemplateRepositoryConnectionParameters
// Send template repository connection parameters to Worker
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendTemplateRepositoryConnectionParameters() {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "972a1a70-3e6c-4f29-98d3-22f5007b409d",
	}).Debug("Incoming 'SendTemplateRepositoryConnectionParameters'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "a1e08cd0-e15d-4ca7-885f-62fa17034763",
	}).Debug("Outgoing 'SendTemplateRepositoryConnectionParameters'")

	var err error

	// Do call-back to get all 	// Create supported TestInstructions, TestInstructionContainers and Allowed Users
	var templateRepositoryConnectionParameters *common_config.RepositoryTemplatePathStruct
	templateRepositoryConnectionParameters = common_config.ConnectorFunctionsToDoCallBackOn.GenerateTemplateRepositoryConnectionParameters()

	// Convert into gRPC-message by looping incomming message
	var templateRepositoryConnectionParametersAsGrpc *fenixExecutionWorkerGrpcApi.AllTemplateRepositoryConnectionParameters
	for messageIndex, repositoryConnectionParameters := range templateRepositoryConnectionParameters.TemplatePaths {

		// Create one url to a Template repository
		var tempAllTemplateRepositories *fenixExecutionWorkerGrpcApi.TemplateRepositoryConnectionParameters
		tempAllTemplateRepositories = &fenixExecutionWorkerGrpcApi.TemplateRepositoryConnectionParameters{
			RepositoryApiUrl: repositoryConnectionParameters.RepositoryApiUrl,
			RepositoryOwner:  repositoryConnectionParameters.RepositoryOwner,
			RepositoryName:   repositoryConnectionParameters.RepositoryName,
			RepositoryPath:   repositoryConnectionParameters.RepositoryPath,
			GitHubApiKey:     common_config.GitHubApiKeys[messageIndex],
		}

		// Add it to the gRPC-message
		templateRepositoryConnectionParametersAsGrpc.AllTemplateRepositories = append(
			templateRepositoryConnectionParametersAsGrpc.GetAllTemplateRepositories(),
			tempAllTemplateRepositories)
	}

	// Add Domain-information
	var tempClientSystemIdentificationMessage *fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage
	tempClientSystemIdentificationMessage = &fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage{
		DomainUuid:          common_config.ThisDomainsUuid,
		ExecutionDomainUuid: common_config.ThisExecutionDomainUuid,
		ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(
			common_config.GetHighestExecutionWorkerProtoFileVersion()),
	}

	templateRepositoryConnectionParametersAsGrpc.ClientSystemIdentification = tempClientSystemIdentificationMessage

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
			"ID":    "00e1967c-ee39-4f22-8e63-fb54ac97fb0a",
			"error": err,
		}).Fatalln("Problem setting up connection to Fenix Execution Worker for 'SendTemplateRepositoryConnectionParameters'")
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
				"ID":                  "5a4d988f-d32d-4547-9a21-76aca8eafa60",
				"returnMessageString": returnMessageString,
			}).Fatalln("Problem generating GCP access token for 'SendTemplateRepositoryConnectionParameters'")
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
			ConnectorPublishTemplateRepositoryConnectionParameters(
				ctx,
				templateRepositoryConnectionParametersAsGrpc)

		// Add to counter for how many gRPC-call-attempts to Worker that have been done
		gRPCCallAttemptCounter = gRPCCallAttemptCounter + 1

		// Shouldn't happen
		if err != nil {

			// Only return the error after last attempt
			if gRPCCallAttemptCounter >= numberOfgRPCCallAttempts {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":    "74336d26-d0ee-47b6-bc34-ff7331116953",
					"error": err,
				}).Fatalln("Problem to do gRPC-call to Fenix Execution Worker for 'SendTemplateRepositoryConnectionParameters'")

			}

			// Sleep for some time before retrying to connect
			time.Sleep(time.Millisecond * time.Duration(sleepTimeBetweenGrpcCallAttempts[gRPCCallAttemptCounter-1]))

		} else if returnMessage.AckNack == false {
			// Couldn't handle gPRC call
			common_config.Logger.WithFields(logrus.Fields{
				"ID":                        "49053c3b-5561-4724-a506-7f655e6b5b65",
				"Message from Fenix Worker": returnMessage.Comments,
			}).Fatalln("Problem to do gRPC-call to Worker for 'SendTemplateRepositoryConnectionParameters'")

		} else {

			common_config.Logger.WithFields(logrus.Fields{
				"ID": "b7642df4-234e-497b-a725-a337582a238e",
			}).Debug("Success in doing gRPC-call to Worker for 'SendTemplateRepositoryConnectionParameters")

			return

		}

	}
}
