package gRPCServer

import (
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/messagesToExecutionWorkerServer"
	fenixExecutionConnectorGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionConnectorGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// AreYouAlive - *********************************************************************
// Anyone can check if Fenix Execution Worker server is alive with this service, should be used to check serves for Connector
func (fenixConnectorGrpcObject *FenixConnectorGrpcServicesServerStruct) WorkerAreYouAlive(ctx context.Context, emptyParameter *fenixExecutionConnectorGrpcApi.EmptyParameter) (*fenixExecutionConnectorGrpcApi.AckNackResponse, error) {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "ee52e7e1-1e2c-47e7-9d9e-36e3229627f3",
	}).Debug("Incoming 'gRPCServer - WorkerAreYouAlive'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "fe41e598-6fdd-4c1f-86e6-463d07db4a3a",
	}).Debug("Outgoing 'gRPCServer - WorkerAreYouAlive'")

	// Don't call Worker if it is not allowed
	if common_config.TurnOffAllCommunicationWithWorker == true {

		ackNackResponseMessage := &fenixExecutionConnectorGrpcApi.AckNackResponse{
			AckNack:                         false,
			Comments:                        "Message to Worker is now allowed due to environment variable: 'TurnOffAllCommunicationWithWorker'=true",
			ErrorCodes:                      nil,
			ProtoFileVersionUsedByConnector: fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(common_config.GetHighestConnectorProtoFileVersion()),
		}

		return ackNackResponseMessage, nil
	}

	// Current user
	userID := "gRPC-api doesn't support UserId"

	// Check if Client is using correct proto files version
	returnMessage := common_config.IsCallerUsingCorrectConnectorProtoFileVersion(userID, emptyParameter.ProtoFileVersionUsedByCaller)
	if returnMessage != nil {

		// Exiting
		return returnMessage, nil
	}

	// Set up instance to use for execution gPRC
	var fenixExecutionWorkerObject *messagesToExecutionWorkerServer.MessagesToExecutionWorkerObjectStruct
	fenixExecutionWorkerObject = &messagesToExecutionWorkerServer.MessagesToExecutionWorkerObjectStruct{}

	response, responseMessage := fenixExecutionWorkerObject.SendAreYouAliveToFenixExecutionServer()

	// Create Error Codes
	var errorCodes []fenixExecutionConnectorGrpcApi.ErrorCodesEnum

	ackNackResponseMessage := &fenixExecutionConnectorGrpcApi.AckNackResponse{
		AckNack:                         response,
		Comments:                        fmt.Sprintf("The response from Worker is '%s'", responseMessage),
		ErrorCodes:                      errorCodes,
		ProtoFileVersionUsedByConnector: fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(common_config.GetHighestConnectorProtoFileVersion()),
	}

	return ackNackResponseMessage, nil

}
