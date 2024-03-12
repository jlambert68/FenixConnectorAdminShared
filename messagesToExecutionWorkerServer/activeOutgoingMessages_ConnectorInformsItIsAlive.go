package messagesToExecutionWorkerServer

import (
	"context"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"time"
)

// SendConnectorInformsItIsAlive - Inform  Worker that Connector is up and running
func (toExecutionWorkerObject *MessagesToExecutionWorkerObjectStruct) SendConnectorInformsItIsAlive(
	connectorIsReadyMessage *fenixExecutionWorkerGrpcApi.ConnectorIsReadyMessage) (err error) {

	/*
		common_config.Logger.WithFields(logrus.Fields{
			"id": "dc761d3f-f85f-4b0e-a06d-755e9b8dd352",
		}).Debug("Incoming 'SendConnectorInformsItIsAlive'")

		common_config.Logger.WithFields(logrus.Fields{
			"id": "a682cce6-4e88-4613-8d14-f579c994b4bf",
		}).Debug("Outgoing 'SendConnectorInformsItIsAlive'")
	*/

	// Before exiting
	defer func() {

	}()

	var ctx context.Context
	var returnMessageAckNack bool

	ctx = context.Background()

	// Set up connection to Server
	ctx, err = toExecutionWorkerObject.SetConnectionToFenixExecutionWorkerServer(ctx)
	if err != nil {
		return err
	}

	// Do gRPC-call
	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		/*
			toExecutionWorkerObject.Logger.WithFields(logrus.Fields{
				"ID": "c6fdb82a-6956-4943-b08c-1f6b5164531f",
			}).Debug("Running Defer Cancel function")
		*/
		cancel()

	}()

	// Only add access token when run on GCP
	if common_config.ExecutionLocationForFenixExecutionWorkerServer == common_config.GCP &&
		common_config.GCPAuthentication == true {

		// Add Access token
		ctx, returnMessageAckNack, _ = gcp.Gcp.GenerateGCPAccessToken(
			ctx, gcp.GetTokenForGrpcAndPubSub)
		if returnMessageAckNack == false {
			return err
		}

	}

	// Do the gRPC-call
	//md2 := MetadataFromHeaders(headers)
	//myctx := metadata.NewOutgoingContext(ctx, md2)

	// Creates a new temporary client only to be used for this call
	var tempFenixExecutionWorkerConnectorGrpcServicesClient fenixExecutionWorkerGrpcApi.FenixExecutionWorkerConnectorGrpcServicesClient
	tempFenixExecutionWorkerConnectorGrpcServicesClient = fenixExecutionWorkerGrpcApi.
		NewFenixExecutionWorkerConnectorGrpcServicesClient(remoteFenixExecutionWorkerServerConnection)

	var connectorIsReadyResponseMessage *fenixExecutionWorkerGrpcApi.ConnectorIsReadyResponseMessage
	connectorIsReadyResponseMessage, err = tempFenixExecutionWorkerConnectorGrpcServicesClient.ConnectorInformsItIsAlive(
		ctx, connectorIsReadyMessage)

	// Shouldn't happen
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":    "41cc0850-93c2-4e57-8baf-11144840e601",
			"error": err,
		}).Error("Problem to do gRPC-call to FenixExecutionWorker for 'SendConnectorInformsItIsAlive'")

		return err

	} else if connectorIsReadyResponseMessage.AckNackResponse.AckNack == false {
		// FenixTestDataSyncServer couldn't handle gPRC call
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                                  "6fcf35a5-6a8f-4b3c-a2a0-e00c9d594c73",
			"Message from Fenix Execution Server": connectorIsReadyResponseMessage.AckNackResponse.Comments,
		}).Error("Problem to do gRPC-call to FenixExecutionWorker for 'SendConnectorInformsItIsAlive'")

		return err
	}

	return err

}
