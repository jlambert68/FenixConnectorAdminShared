package fenixConnectorAdminShared

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
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

	// When Execution Worker runs on GCP, then set up access
	if common_config.ExecutionLocationForFenixExecutionWorkerServer == common_config.GCP &&
		common_config.GCPAuthentication == true &&
		common_config.TurnOffCallToWorker == false {
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

	// Start Connector Engine
	fenixExecutionConnectorMain()

	/*

		// Run as console program and exit as on standard exiting signals
		sig := make(chan os.Signal, 1)
		done := make(chan bool, 1)

		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sig
			fmt.Println()
			fmt.Println(sig)
			done <- true

			fmt.Println("ctrl+c")
		}()

		fmt.Println("awaiting signal")
		<-done
		fmt.Println("exiting")

	*/

}
