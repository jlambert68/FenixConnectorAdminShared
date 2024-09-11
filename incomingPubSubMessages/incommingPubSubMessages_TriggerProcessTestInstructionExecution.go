package incomingPubSubMessages

import (
	"errors"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/connectorEngine"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

func triggerProcessTestInstructionExecution(pubSubMessage []byte) (err error) {

	// Remove any unwanted characters
	// Remove '\n'
	var cleanedMessage string
	var cleanedMessageAsByteArray []byte
	var pubSubMessageAsString string

	pubSubMessageAsString = string(pubSubMessage)
	//cleanedMessage = strings.ReplaceAll(pubSubMessageAsString, "\n", "")

	// Replace '\"' with '"'
	//cleanedMessage = strings.ReplaceAll(cleanedMessage, "\\\"", "\"")

	//cleanedMessage = strings.ReplaceAll(cleanedMessage, " ", "")

	cleanedMessage = pubSubMessageAsString

	// Convert back into byte-array
	cleanedMessageAsByteArray = []byte(cleanedMessage)

	// Convert PubSub-message back into proto-message
	var processTestInstructionExecutionPubSubRequest fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest
	err = protojson.Unmarshal(cleanedMessageAsByteArray, &processTestInstructionExecutionPubSubRequest)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":                         "bb8e4c1c-12d9-4d19-b77c-165dd05fd4eb",
			"Error":                      err,
			"string(pubSubMessage.Data)": string(pubSubMessage),
		}).Error("Something went wrong when converting 'PubSub-message into proto-message")

		// Drop this message, without sending 'Ack'
		return err
	}

	// Check if this message has already been processed
	var uniqueMessageIdentifier string
	var isDuplicate bool

	uniqueMessageIdentifier = processTestInstructionExecutionPubSubRequest.GetTestInstruction().
		GetTestInstructionExecutionUuid() +
		strconv.Itoa(int(processTestInstructionExecutionPubSubRequest.GetTestInstruction().
			GetTestInstructionExecutionVersion()))

	isDuplicate, err = checkForDuplicatesAndAddForMessages(uniqueMessageIdentifier)

	// If there was an error when checking for duplicate then return back for next message
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":                         "f315bb41-8d09-4e78-973d-0136cd93e5b7",
			"Error":                      err,
			"string(pubSubMessage.Data)": string(pubSubMessage),
		}).Error("Got some problem when checking if this message is a duplicate")

		return err
	}

	// If it is a duplicate message then just return back for next message
	if isDuplicate == true {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":                         "31ff2119-3b12-45d8-999c-3b27b5c47894",
			"string(pubSubMessage.Data)": string(pubSubMessage),
		}).Error("Message is a duplicate")

		return nil
	}

	var couldSend bool
	var returnMessage string

	// Gets the max time for when the TestInstructionExecution can be seen as "dead", by doing callback to code in Connector
	var maxExpectedFinishedTimeStamp time.Time
	var processTestInstructionExecutionResponse *fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionResponse
	maxExpectedFinishedTimeStamp, err = common_config.ConnectorFunctionsToDoCallBackOn.GetMaxExpectedFinishedTimeStamp(
		&processTestInstructionExecutionPubSubRequest)

	// Response from GetMaxExpectedFinishedTimeStamp-call
	if err != nil {

		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "8b177a17-efe9-4041-89a9-bd9b424509d9",
			"err": err.Error(),
		}).Error("Couldn't get a ExpectedFinishTimeStamp")

		// Couldn't get a ExpectedFinishTimeStamp
		timeAtDurationEnd := time.Now()

		// Generate response message to Worker, that conversion didn't work out
		processTestInstructionExecutionResponse = &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionResponse{
			AckNackResponse: &fenixExecutionWorkerGrpcApi.AckNackResponse{
				AckNack:                      false,
				Comments:                     err.Error(),
				ErrorCodes:                   nil,
				ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(common_config.GetHighestExecutionWorkerProtoFileVersion()),
			},
			TestInstructionExecutionUuid:   processTestInstructionExecutionPubSubRequest.TestInstruction.TestInstructionExecutionUuid,
			ExpectedExecutionDuration:      timestamppb.New(timeAtDurationEnd),
			TestInstructionCanBeReExecuted: true,
		}

	} else {

		// Got an OK MaxExpectedFinishedTimeStamp, so generate OK response message to Worker
		processTestInstructionExecutionResponse = &fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionResponse{
			AckNackResponse: &fenixExecutionWorkerGrpcApi.AckNackResponse{
				AckNack:                      true,
				Comments:                     "",
				ErrorCodes:                   nil,
				ProtoFileVersionUsedByClient: fenixExecutionWorkerGrpcApi.CurrentFenixExecutionWorkerProtoFileVersionEnum(common_config.GetHighestExecutionWorkerProtoFileVersion()),
			},
			TestInstructionExecutionUuid:   processTestInstructionExecutionPubSubRequest.TestInstruction.TestInstructionExecutionUuid,
			ExpectedExecutionDuration:      timestamppb.New(maxExpectedFinishedTimeStamp),
			TestInstructionCanBeReExecuted: false,
		}
	}

	// Send 'ProcessTestInstructionExecutionPubSubRequest-response' back to worker over direct gRPC-call
	couldSend, returnMessage = connectorEngine.TestInstructionExecutionEngine.
		MessagesToExecutionWorkerObjectReference.
		SendConnectorProcessTestInstructionExecutionResponse(processTestInstructionExecutionResponse)

	if couldSend == false {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":            "34f286c7-24c2-480e-8f57-adea8b96380c",
			"returnMessage": returnMessage,
		}).Error("Couldn't send response to Worker")

		// Drop this message, without sending 'Ack'
		err = errors.New(returnMessage)
		return err

	} else {

		// Send TestInstruction to Connector via call-back
		var finalTestInstructionExecutionResultMessage *fenixExecutionWorkerGrpcApi.FinalTestInstructionExecutionResultMessage
		finalTestInstructionExecutionResultMessage, err = common_config.ConnectorFunctionsToDoCallBackOn.
			ProcessTestInstructionExecutionRequest(&processTestInstructionExecutionPubSubRequest)

		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"ID":  "60598841-9c08-410c-ad63-c8ac7ee10db4",
				"err": err.Error(),
			}).Error("Got some error when sending TestInstruction for processing by Connector via call-back")
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

		// Send 'FinalTestInstructionExecutionResultMessage' back to worker over direct gRPC-call
		couldSend, returnMessage = connectorEngine.TestInstructionExecutionEngine.MessagesToExecutionWorkerObjectReference.
			SendReportCompleteTestInstructionExecutionResultToFenixWorkerServer(finalTestInstructionExecutionResultMessage)

		if couldSend == false {
			common_config.Logger.WithFields(logrus.Fields{
				"ID": "1ce93ee2-5542-4437-9c05-d7f9d19313fa",
				"finalTestInstructionExecutionResultMessage": finalTestInstructionExecutionResultMessage,
				"returnMessage": returnMessage,
			}).Error("Couldn't send response to Worker")

			err = errors.New(returnMessage)
			return err

		} else {

			// Send 'Ack' back to PubSub-system that message has taken care of
			return nil
		}

	}

}
