package incomingPubSubMessages

import (
	"context"
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

// PullPubSubTestInstructionExecutionMessagesGcpRestApi
// Use GCP RestApi to subscribe to a PubSub-Topic
func PullPubSubTestInstructionExecutionMessagesGcpRestApi(connectorIsReadyToReceiveWorkChannelPtr *chan bool) {

	connectorIsReadyToReceiveWorkChannel := *connectorIsReadyToReceiveWorkChannelPtr

	common_config.Logger.WithFields(logrus.Fields{
		"id": "e2695a29-3412-48ff-ab51-662c711fef44",
	}).Debug("Incoming 'PullPubSubTestInstructionExecutionMessagesGcpRestApi'")

	defer common_config.Logger.WithFields(logrus.Fields{
		"id": "e61fd7f6-95dd-4bbc-a7ae-ee8c4571174f",
	}).Debug("Outgoing 'PullPubSubTestInstructionExecutionMessagesGcpRestApi'")

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
	/*
		// Add Access token via GCP login
		var returnMessageAckNack bool
		var returnMessageString string

		// When Connector is NOT running in GCP
		if common_config.ExecutionLocationForConnector != common_config.GCP {
			_, returnMessageAckNack, returnMessageString = gcp.Gcp.GenerateGCPAccessToken(context.Background(), gcp.GetTokenForGrpcAndPubSub) //gcp.GetTokenFromWorkerForPubSub) //gcp.GenerateTokenForPubSub)
			if returnMessageAckNack == false {

				common_config.Logger.WithFields(logrus.Fields{
					"ID":                   "c96f4e4a-93b8-4175-ac2e-5b4eacd89a8f",
					"returnMessageAckNack": returnMessageAckNack,
					"returnMessageString":  returnMessageString,
				}).Error("Got some problem when generating GCP access token")

				return
			}
		}
	*/

	// Generate Subscription name to use
	var subID string
	subID = generatePubSubTopicSubscriptionNameForExecutionStatusUpdates()

	// Create a loop to be able to have a continuous PubSub Subscription Engine
	var numberOfMessagesInPullResponse int
	var err error
	var returnAckNack bool
	var returnMessage string
	var ctx context.Context
	var accessToken string

	resetTickerCondition := true
	falseCount := 0
	currentInterval := time.Second

	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	ctx = context.Background()

	for range ticker.C {

		// Generate a new token is needed
		_, returnAckNack, returnMessage = gcp.Gcp.GenerateGCPAccessToken(ctx, gcp.GenerateTokenForPubSub)
		if returnAckNack == false {

			// Set to zero because we need some waiting time
			numberOfMessagesInPullResponse = 0

			common_config.Logger.WithFields(logrus.Fields{
				"id":            "4d4f1144-a905-4b3c-8d71-ef533eea514c",
				"returnMessage": returnMessage,
			}).Error("Problem when generating a new token. Waiting some time before next try")

		} else {

			fmt.Printf("Tick at: %v, numberOfMessagesInPullResponse: %d, Interval: %v\n", time.Now(), numberOfMessagesInPullResponse, currentInterval)
			if resetTickerCondition == true {

				// Reset the false count and interval when resetTickerCondition is true
				falseCount = 0
				if currentInterval != time.Second {
					currentInterval = time.Second
					ticker.Reset(currentInterval)
				}

			} else {

				// Increase falseCount and, when reached max tries, then ramp up Ticker delay
				falseCount++
				if falseCount >= 10 {
					// Ramp up the interval slowly up to 60 seconds
					if currentInterval < 60*time.Second {
						newInterval := nextInterval(currentInterval)
						if newInterval != currentInterval {
							currentInterval = newInterval
							ticker.Reset(currentInterval)
							fmt.Printf("Interval adjusted to: %v\n", currentInterval)
						}
					}
				}
			}

			// Get AccessToken
			accessToken, err = gcp.Gcp.GetGcpAccessTokenForAuthorizedAccountsPubSub()
			if err != nil {
				common_config.Logger.WithFields(logrus.Fields{
					"ID":  "3480675a-e306-4bec-9e3a-709f20311a5c",
					"err": err,
				}).Error("Got som problem when retrieving access token")

			} else {

				// Pull a certain number of messages from Subscription
				numberOfMessagesInPullResponse, err = retrievePubSubMessagesViaRestApi(subID, accessToken)

				if err != nil {

					common_config.Logger.WithFields(logrus.Fields{
						"ID":          "7efbd7d7-7761-4c94-8306-ac7349cb93c9",
						"accessToken": accessToken,
						"err":         err,
					}).Error("Got som problem when doing PubSub-receive")
				}
			}

			// Reset Ticker condition if there were any messages
			if numberOfMessagesInPullResponse > 0 && err == nil {
				resetTickerCondition = true
			} else {
				resetTickerCondition = false
			}
		}
	}
}

// Calculate the next wait interval before checking PubSub
func nextInterval(current time.Duration) time.Duration {
	const maxDuration = 60 * time.Second
	const minDuration = 1 * time.Second
	const midPoint = 30 * time.Second

	// Calculate how far the current duration is from the midpoint in percentage
	midpointOffset := float64(current-minDuration) / float64(midPoint-minDuration)

	// Calculate a scaling factor using a simple parabolic equation, scaled to slow down initial and final increments
	scaleFactor := 4 * (midpointOffset - math.Pow(midpointOffset, 2)) // This creates a parabolic curve

	// Determine the increment, scaled by the calculated factor
	increment := time.Duration(float64(maxDuration-minDuration) * scaleFactor / 10)
	if increment < time.Second {
		increment = time.Second
	}

	newInterval := current + increment
	if newInterval > maxDuration {
		newInterval = maxDuration
	}

	return newInterval
}
