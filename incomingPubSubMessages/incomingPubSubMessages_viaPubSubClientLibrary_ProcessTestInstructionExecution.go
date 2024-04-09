package incomingPubSubMessages

import (
	"cloud.google.com/go/pubsub"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
)

// PullPubSubTestInstructionExecutionMessagesGcpClientLib
// Use GCP Client Library to subscribe to a PubSub-Topic
func PullPubSubTestInstructionExecutionMessagesGcpClientLib(connectorIsReadyToReceiveWorkChannelPtr *chan bool) {

	connectorIsReadyToReceiveWorkChannel := *connectorIsReadyToReceiveWorkChannelPtr

	common_config.Logger.WithFields(logrus.Fields{
		"id": "1ed3f12b-65fb-41bc-b12d-ca4af21e8a36",
	}).Debug("Incoming 'PullPubSubTestInstructionExecutionMessagesGcpClientLib'")

	defer common_config.Logger.WithFields(logrus.Fields{
		"id": "1a6fc28d-2523-44c7-9f02-8a7392c4966c",
	}).Debug("Outgoing 'PullPubSubTestInstructionExecutionMessagesGcpClientLib'")

	// Before Starting PubSub-receiver secure that an access token has been received
	for {
		var responseFromWorkerReceived bool
		responseFromWorkerReceived = <-connectorIsReadyToReceiveWorkChannel

		if responseFromWorkerReceived == true {
			// Continue when we got an access token
			break
		} else {

		}

	}
	projectID := common_config.GcpProject
	subID := generatePubSubTopicSubscriptionNameForExecutionStatusUpdates()

	var pubSubClient *pubsub.Client
	var err error
	var opts []grpc.DialOption
	var appendedCtx context.Context
	var returnAckNack bool
	var returnMessage string

	ctx := context.Background()

	// Add Access token
	//var returnMessageAckNack bool
	//var returnMessageString string

	//ctx, returnMessageAckNack, returnMessageString = gcp.Gcp.GenerateGCPAccessToken(ctx, gcp.GenerateTokenForPubSub)
	//if returnMessageAckNack == false {
	//	return errors.New(returnMessageString)
	//}

	if len(common_config.LocalServiceAccountPath) != 0 {
		//ctx = context.Background()
		pubSubClient, err = pubsub.NewClient(ctx, projectID)
	} else {

	}
	//When running on GCP then use credential otherwise not
	if true { //common_config.ExecutionLocationForWorker == common_config.GCP {

		var creds credentials.TransportCredentials
		creds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})

		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
		}

		// Should ProxyServer be used for outgoing connections
		if common_config.ShouldProxyServerBeUsed == true {
			// Use Proxy
			remoteFenixExecutionWorkerServerConnection, err := gcp.GRPCDialer(
				"",
				common_config.FenixExecutionWorkerAddress,
				common_config.FenixExecutionWorkerPort)

			if err != nil {
				common_config.Logger.WithFields(logrus.Fields{
					"ID":            "a94c46ba-87df-4595-96d5-e4d144f7bc2c",
					"error message": err,
				}).Error("Couldn't generate connection to PubSub-server via Proxy Server")

				return
			}

			appendedCtx, returnAckNack, returnMessage = gcp.Gcp.GenerateGCPAccessToken(ctx, gcp.GetTokenForGrpcAndPubSub)
			if returnAckNack == false {
				err = errors.New(returnMessage)
				return
			}

			pubSubClient, err = pubsub.NewClient(appendedCtx, projectID, option.WithGRPCConn(remoteFenixExecutionWorkerServerConnection))

		} else {

			pubSubClient, err = pubsub.NewClient(ctx, projectID, option.WithGRPCDialOption(opts[0]))
		}

	}

	if err != nil {

		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "a606a500-7e76-43a5-876d-c51ff8a683ec",
			"err": err,
		}).Error("Got some problem when creating 'pubsub.NewClient'")

		return
	}
	defer pubSubClient.Close()

	clientSubscription := pubSubClient.Subscription(subID)

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	//ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	//defer cancel()

	err = clientSubscription.Receive(ctx, func(_ context.Context, pubSubMessage *pubsub.Message) {

		common_config.Logger.WithFields(logrus.Fields{
			"ID": "8e75e797-7d75-45fa-93b2-1190c48dd0af",
		}).Debug(fmt.Printf("Got message: %q", string(pubSubMessage.Data)))

		// Remove any unwanted characters
		// Remove '\n'
		var cleanedMessage string
		var cleanedMessageAsByteArray []byte
		var pubSubMessageAsString string

		pubSubMessageAsString = string(pubSubMessage.Data)
		cleanedMessage = strings.ReplaceAll(pubSubMessageAsString, "\n", "")

		// Replace '\"' with '"'
		cleanedMessage = strings.ReplaceAll(cleanedMessage, "\\\"", "\"")

		cleanedMessage = strings.ReplaceAll(cleanedMessage, " ", "")

		// Convert back into byte-array
		cleanedMessageAsByteArray = []byte(cleanedMessage)

		// Convert PubSub-message back into proto-message
		var processTestInstructionExecutionPubSubRequest fenixExecutionWorkerGrpcApi.ProcessTestInstructionExecutionPubSubRequest
		err = protojson.Unmarshal(cleanedMessageAsByteArray, &processTestInstructionExecutionPubSubRequest)
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"Id":                         "df899f00-5e6a-49c3-a272-ecb3e4bc19f2",
				"Error":                      err,
				"string(pubSubMessage.Data)": string(pubSubMessage.Data),
			}).Error("Something went wrong when converting 'PubSub-message into proto-message")

			// Drop this message, without sending 'Ack'
			return
		}

		// Trigger TestInstruction in parallel while processing next message
		go func() {

			/*
				// Channel to signal when message processing is done
				doneProcessing := make(chan bool)

				// Start a goroutine to extend the ack deadline
				go func() {
					ticker := time.NewTicker(25 * time.Second)
					defer ticker.Stop()

					for {
						select {
						case <-ticker.C:
							// Extend the deadline by 30 seconds
							subConfigToUpdate := pubsub.SubscriptionConfigToUpdate{
								AckDeadline: 30 * time.Second, // changing the ack deadline
							}

							// Perform the update
							_, err = clientSubscription.Update(ctx, subConfigToUpdate)
							if err != nil {

								common_config.Logger.WithFields(logrus.Fields{
									"ID":                 "476cc867-8fc7-40ef-85c7-fc959d397003",
									"pubSubMessage.Data": string(pubSubMessage.Data),
								}).Error("Couldn't update 'AckDeadline'")

								return
							}

						case <-doneProcessing:
							return
						}
					}
				}()

			*/

			// Acknowledge the message
			// Send 'Ack' back to PubSub-system that message has taken care of, even without a successful execution
			pubSubMessage.Ack()

			err = triggerProcessTestInstructionExecution(pubSubMessage.Data)

			// stop prolonging of 'AckDeadline'
			//doneProcessing <- true

			if err == nil {

				// Acknowledge the message
				// Send 'Ack' back to PubSub-system that message has taken care of
				//pubSubMessage.Ack()

			} else {

				common_config.Logger.WithFields(logrus.Fields{
					"ID": "2d74199d-a434-4658-a085-46a83c14c8fb",
				}).Error("Failed to Process TestInstructionExecution")

			}
		}()

	})

	// Shouldn't happen
	if err != nil {

		common_config.Logger.WithFields(logrus.Fields{
			"ID":  "8acdd7af-d71e-41e0-a7e0-0b36c79c952f",
			"err": err,
		}).Fatalln("PubSub receiver for TestInstructionExecutions ended, which is not intended")

	}

}
