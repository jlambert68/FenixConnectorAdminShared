package gRPCServer

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	fenixExecutionConnectorGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionConnectorGrpcApi/go_grpc_api"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// TriggerTestInstructionExecution - *********************************************************************
// Trigger Connector to Execute a TestInstruction, used for testing both internal execution code and when calls are
// done towards external execution. Can use local Test WebServer, internal execution logic or external execution logic
func (fenixConnectorGrpcObject *FenixConnectorGrpcServicesServerStruct) TriggerTestInstructionExecution(
	ctx context.Context,
	processTestInstructionExecutionPubSubRequest *fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest) (
	*fenixExecutionConnectorGrpcApi.AckNackResponse, error) {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "aae71a08-8647-40e5-b2fe-02904b3eb11f",
	}).Debug("Incoming 'gRPCServer - processTestInstructionExecutionPubSubRequest'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "8fc14487-e148-473e-9711-f25399a9da5d",
	}).Debug("Outgoing 'gRPCServer - processTestInstructionExecutionPubSubRequest'")

	// Current user
	userID := "gRPC-api doesn't support UserId"

	// Check if Client is using correct proto files version
	returnMessage := common_config.IsCallerUsingCorrectConnectorProtoFileVersion(
		userID,
		fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(
			processTestInstructionExecutionPubSubRequest.GetDomainIdentificationAnfProtoFileVersionUsedByClient().
				GetProtoFileVersionUsedByClient()))
	if returnMessage != nil {

		// Exiting
		return returnMessage, nil
	}

	// Send TestInstruction to Connector to be executed, via call-back
	var err error
	var finalTestInstructionExecutionResultMessage *fenixExecutionWorkerGrpcApi.FinalTestInstructionExecutionResultMessage
	finalTestInstructionExecutionResultMessage, err = common_config.ConnectorFunctionsToDoCallBackOn.
		ProcessTestInstructionExecutionRequest(processTestInstructionExecutionPubSubRequest)

	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "e01f9bdb-b2b9-4389-9be3-c1c3fdc36e46",
			"err": err.Error(),
		}).Error("Got some error when sending TestInstruction for processing by Connector via call-back")

		ackNackResponseMessage := &fenixExecutionConnectorGrpcApi.AckNackResponse{
			AckNack:    false,
			Comments:   err.Error(),
			ErrorCodes: nil,
			ProtoFileVersionUsedByConnector: fenixExecutionConnectorGrpcApi.
				CurrentFenixExecutionConnectorProtoFileVersionEnum(common_config.GetHighestConnectorProtoFileVersion()),
		}

		return ackNackResponseMessage, nil
	}

	// Add Domain-information
	var tempClientSystemIdentificationMessage *fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage
	tempClientSystemIdentificationMessage = &fenixExecutionWorkerGrpcApi.ClientSystemIdentificationMessage{
		DomainUuid:          common_config.ThisDomainsUuid,
		ExecutionDomainUuid: common_config.ThisExecutionDomainUuid,
		ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(
			common_config.GetHighestExecutionWorkerProtoFileVersion()),
	}
	finalTestInstructionExecutionResultMessage.ClientSystemIdentification = tempClientSystemIdentificationMessage

	// Log 'finalTestInstructionExecutionResultMessage'
	common_config.Logger.WithFields(logrus.Fields{
		"ID": "58dd2ef8-1ed9-4d15-ba66-61e01df7f911",
		"finalTestInstructionExecutionResultMessage": finalTestInstructionExecutionResultMessage,
	}).Info("Success in send TestInstruction for execution")

	// Create Error Codes
	var errorCodes []fenixExecutionConnectorGrpcApi.ErrorCodesEnum

	ackNackResponseMessage := &fenixExecutionConnectorGrpcApi.AckNackResponse{
		AckNack:                         true,
		Comments:                        "Success in send TestInstruction for Execution. Look in log for mor details.",
		ErrorCodes:                      errorCodes,
		ProtoFileVersionUsedByConnector: fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(common_config.GetHighestConnectorProtoFileVersion()),
	}

	return ackNackResponseMessage, nil

}
