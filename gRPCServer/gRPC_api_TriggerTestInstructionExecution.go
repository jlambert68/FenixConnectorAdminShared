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
	triggerProcessTestInstructionExecutionPubSubRequest *fenixExecutionConnectorGrpcApi.TriggerProcessTestInstructionExecutionPubSubRequest) (
	*fenixExecutionConnectorGrpcApi.AckNackResponse, error) {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "aae71a08-8647-40e5-b2fe-02904b3eb11f",
	}).Debug("Incoming 'gRPCServer - processTestInstructionExecutionPubSubRequest'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "8fc14487-e148-473e-9711-f25399a9da5d",
	}).Debug("Outgoing 'gRPCServer - processTestInstructionExecutionPubSubRequest'")

	/*
		// Current user
		userID := "gRPC-api doesn't support UserId"

		// Check if Client is using correct proto files version
		returnMessage := common_config.IsCallerUsingCorrectConnectorProtoFileVersion(
			userID,
			fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(
				triggerProcessTestInstructionExecutionPubSubRequest.GetDomainIdentificationAnfProtoFileVersionUsedByClient().
					GetProtoFileVersionUsedByClient()))
		if returnMessage != nil {

			// Exiting
			return returnMessage, nil
		}
	*/

	// Convert 'triggerProcessTestInstructionExecutionPubSubRequest' into 'processTestInstructionExecutionPubSubRequest'
	var processTestInstructionExecutionPubSubRequest *fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest
	processTestInstructionExecutionPubSubRequest = &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest{
		DomainIdentificationAnfProtoFileVersionUsedByClient: &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest_ClientSystemIdentificationMessage{
			DomainUuid: triggerProcessTestInstructionExecutionPubSubRequest.
				GetDomainIdentificationAnfProtoFileVersionUsedByClient().GetDomainUuid(),
			ExecutionDomainUuid: triggerProcessTestInstructionExecutionPubSubRequest.
				GetDomainIdentificationAnfProtoFileVersionUsedByClient().GetExecutionDomainUuid(),
			ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.
				ProcessTestInstructionExecutionPubSubRequest_CurrentFenixExecutionWorkerProtoFileVersionEnum(
					triggerProcessTestInstructionExecutionPubSubRequest.
						GetDomainIdentificationAnfProtoFileVersionUsedByClient().GetProtoFileVersionUsedByClient()),
		},
		TestInstruction: &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest_TestInstructionExecutionMessage{
			TestInstructionExecutionUuid: triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
				GetTestInstructionExecutionUuid(),
			TestInstructionUuid: triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
				GetTestInstructionUuid(),
			TestInstructionName: triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
				GetTestInstructionName(),
			MajorVersionNumber: triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
				GetMajorVersionNumber(),
			MinorVersionNumber: triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
				GetMinorVersionNumber(),
			TestInstructionAttributes: nil,
		},
		TestCaseExecutionUuid: triggerProcessTestInstructionExecutionPubSubRequest.GetTestCaseExecutionUuid(),
		TestData: &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest_TestDataMessage{
			TestDataSetUuid: triggerProcessTestInstructionExecutionPubSubRequest.GetTestData().
				GetTestDataSetUuid(),
			ManualOverrideForTestData: nil,
		},
	}

	// Convert 'TestInstructionAttributes' and add
	var tempTestInstructionAttributes []*fenixExecutionWorkerGrpcApi.
		ProcessTestInstructionExecutionPubSubRequest_TestInstructionAttributeMessage
	for _, testInstructionAttribute := range triggerProcessTestInstructionExecutionPubSubRequest.GetTestInstruction().
		GetTestInstructionAttributes() {

		// Create a new Attribute copy
		var tempTestInstructionAttribute *fenixExecutionWorkerGrpcApi.
			ProcessTestInstructionExecutionPubSubRequest_TestInstructionAttributeMessage

		tempTestInstructionAttribute = &fenixExecutionWorkerGrpcApi.
			ProcessTestInstructionExecutionPubSubRequest_TestInstructionAttributeMessage{
			TestInstructionAttributeType: fenixExecutionWorkerGrpcApi.
				ProcessTestInstructionExecutionPubSubRequest_TestInstructionAttributeTypeEnum(
					testInstructionAttribute.TestInstructionAttributeType),
			TestInstructionAttributeUuid:     testInstructionAttribute.TestInstructionAttributeUuid,
			TestInstructionAttributeName:     testInstructionAttribute.TestInstructionAttributeName,
			AttributeValueAsString:           testInstructionAttribute.AttributeValueAsString,
			AttributeValueUuid:               testInstructionAttribute.AttributeValueUuid,
			TestInstructionAttributeTypeUuid: testInstructionAttribute.TestInstructionAttributeTypeUuid,
			TestInstructionAttributeTypeName: testInstructionAttribute.TestInstructionAttributeTypeName,
		}

		// Append Attribute to slice
		tempTestInstructionAttributes = append(tempTestInstructionAttributes, tempTestInstructionAttribute)

	}

	// Add the converted Attributes
	processTestInstructionExecutionPubSubRequest.TestInstruction.TestInstructionAttributes = tempTestInstructionAttributes

	// Convert 'ManualOverrideForTestData' and add
	var tempManualOverrideForTestDataMessages []*fenixExecutionWorkerGrpcApi.
		ProcessTestInstructionExecutionPubSubRequest_TestDataMessage_ManualOverrideForTestDataMessage
	for _, manualOverrideForTestData := range triggerProcessTestInstructionExecutionPubSubRequest.GetTestData().
		GetManualOverrideForTestData() {

		// Create a new ManualOverrideForTestData copy
		var tempManualOverrideForTestDataMessage *fenixExecutionWorkerGrpcApi.
			ProcessTestInstructionExecutionPubSubRequest_TestDataMessage_ManualOverrideForTestDataMessage

		tempManualOverrideForTestDataMessage = &fenixExecutionWorkerGrpcApi.
			ProcessTestInstructionExecutionPubSubRequest_TestDataMessage_ManualOverrideForTestDataMessage{
			TestDataSetAttributeUuid:  manualOverrideForTestData.TestDataSetAttributeUuid,
			TestDataSetAttributeName:  manualOverrideForTestData.TestDataSetAttributeName,
			TestDataSetAttributeValue: manualOverrideForTestData.TestDataSetAttributeValue,
		}

		// Append ManualOverrideForTestData to slice
		tempManualOverrideForTestDataMessages = append(tempManualOverrideForTestDataMessages, tempManualOverrideForTestDataMessage)

	}

	// Add the converted ManualOverrideForTestData
	processTestInstructionExecutionPubSubRequest.TestInstruction.TestInstructionAttributes = tempTestInstructionAttributes

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
