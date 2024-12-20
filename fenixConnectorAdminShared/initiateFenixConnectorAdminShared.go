package fenixConnectorAdminShared

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	"github.com/jlambert68/FenixConnectorAdminShared/privateKeyManager"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"time"
)

func InitiateFenixConnectorAdminShared(
	connectorFunctionsToDoCallBackOn *common_config.ConnectorCallBackFunctionsStruct) {

	// Store references to call-back functions
	common_config.ConnectorFunctionsToDoCallBackOn = connectorFunctionsToDoCallBackOn

	// Run 'init()'
	fenixConnectorAdminSharedInit()

	// Initiate logger in common_config
	InitLogger("")

	// Initiate Logger via Call-back
	common_config.ConnectorFunctionsToDoCallBackOn.InitiateLogger(common_config.Logger)

	//

	// When Execution Worker runs on GCP, then set up access
	if common_config.ExecutionLocationForFenixExecutionWorkerServer == common_config.GCP &&
		common_config.GCPAuthentication == true &&
		common_config.TurnOffAllCommunicationWithWorker == false {
		gcp.Gcp = gcp.GcpObjectStruct{}

		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

		// Generate first time Access token
		_, returnMessageAckNack, returnMessageString := gcp.Gcp.GenerateGCPAccessToken(ctx, gcp.GetTokenForGrpcAndPubSub)
		if returnMessageAckNack == false {

			// If there was any problem then exit program
			common_config.Logger.WithFields(logrus.Fields{
				"id": "20c90d94-eef7-4819-ba8c-b7a56a39f995",
			}).Fatalf("Couldn't generate access token for GCP, return message: '%s'", returnMessageString)

		}
	}

	// Check if new PrivateKey should be generated and stored in secret Manager
	privateKeyManager.ShouldNewPrivateKeyBeGenerated()

	// Start Connector Engine
	fenixExecutionConnectorMain()

}
