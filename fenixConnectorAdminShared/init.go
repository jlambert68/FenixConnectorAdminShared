package fenixConnectorAdminShared

import (
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixSyncShared/environmentVariables"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func fenixConnectorAdminSharedInit() {
	//executionLocationForConnector := flag.String("startupType", "0", "The application should be started with one of the following: LOCALHOST_NODOCKER, LOCALHOST_DOCKER, GCP")
	//flag.Parse()

	var err error

	// Get Environment variable to tell how/were this worker is  running
	var executionLocationForConnector = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionLocationForConnector")

	switch executionLocationForConnector {
	case "LOCALHOST_NODOCKER":
		common_config.ExecutionLocationForConnector = common_config.LocalhostNoDocker

	case "LOCALHOST_DOCKER":
		common_config.ExecutionLocationForConnector = common_config.LocalhostDocker

	case "GCP":
		common_config.ExecutionLocationForConnector = common_config.GCP

	default:
		fmt.Println("Unknown Execution location for Connector: " + executionLocationForConnector + ". Expected one of the following: 'LOCALHOST_NODOCKER', 'LOCALHOST_DOCKER', 'GCP'")
		os.Exit(0)

	}

	// Get Environment variable to tell were Fenix Execution Server is running
	var executionLocationForExecutionWorker = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionLocationForWorker")

	switch executionLocationForExecutionWorker {
	case "LOCALHOST_NODOCKER":
		common_config.ExecutionLocationForFenixExecutionWorkerServer = common_config.LocalhostNoDocker

	case "LOCALHOST_DOCKER":
		common_config.ExecutionLocationForFenixExecutionWorkerServer = common_config.LocalhostDocker

	case "GCP":
		common_config.ExecutionLocationForFenixExecutionWorkerServer = common_config.GCP

	default:
		fmt.Println("Unknown Execution location for Fenix Execution Worker Server: " + executionLocationForExecutionWorker + ". Expected one of the following: 'LOCALHOST_NODOCKER', 'LOCALHOST_DOCKER', 'GCP'")
		os.Exit(0)

	}

	// Address to Fenix Execution Worker Server
	common_config.FenixExecutionWorkerAddress = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionWorkerAddress")

	// Port for Fenix Execution Worker Server
	common_config.FenixExecutionWorkerPort, err = strconv.Atoi(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionWorkerPort"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'ExecutionWorkerPort' to an integer, error: ", err)
		os.Exit(0)

	}

	// Port for Fenix Execution Connector Server
	common_config.ExecutionConnectorPort, err = strconv.Atoi(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionConnectorPort"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'executionConnectorPort' to an integer, error: ", err)
		os.Exit(0)

	}

	// Build the Dial-address for gPRC-call
	common_config.FenixExecutionWorkerAddressToDial = common_config.FenixExecutionWorkerAddress + ":" + strconv.Itoa(common_config.FenixExecutionWorkerPort)

	// Extract Debug level
	var loggingLevel = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("LoggingLevel")

	switch loggingLevel {

	case "DebugLevel":
		common_config.LoggingLevel = logrus.DebugLevel

	case "InfoLevel":
		common_config.LoggingLevel = logrus.InfoLevel

	default:
		fmt.Println("Unknown loggingLevel '" + loggingLevel + "'. Expected one of the following: 'DebugLevel', 'InfoLevel'")
		os.Exit(0)

	}

	// Extract if there is a need for authentication when going toward GCP
	boolValue, err := strconv.ParseBool(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("GCPAuthentication"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'GCPAuthentication:' to an boolean, error: ", err)
		os.Exit(0)
	}
	common_config.GCPAuthentication = boolValue

	// Extract if Service Account should be used towards GCP or should the user log in via web
	boolValue, err = strconv.ParseBool(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("UseServiceAccount"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'UseServiceAccount:' to an boolean, error: ", err)
		os.Exit(0)
	}
	common_config.UseServiceAccount = boolValue

	// Extract OAuth 2.0 Client ID
	common_config.AuthClientId = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("AuthClientId")

	// Extract OAuth 2.0 Client Secret
	common_config.AuthClientSecret = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("AuthClientSecret")

	// Extract the GCP-project
	common_config.GcpProject = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("GcpProject")

	// Extract if PubSub should be used to receive messages from Worker
	common_config.UsePubSubToReceiveMessagesFromWorker, err = strconv.ParseBool(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("UsePubSubToReceiveMessagesFromWorker"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'UsePubSubToReceiveMessagesFromWorker:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract the LocalServiceAccountPath
	common_config.LocalServiceAccountPath = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("LocalServiceAccountPath")
	// The only way have an OK space is to replace an existing character
	if common_config.LocalServiceAccountPath == "#" {
		common_config.LocalServiceAccountPath = ""
	}

	// Set the environment varaible that Google-client-libraries look for
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", common_config.LocalServiceAccountPath)

	// Extract environment variable for 'TestInstructionExecutionPubSubTopicBase'
	common_config.TestInstructionExecutionPubSubTopicBase = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("TestInstructionExecutionPubSubTopicBase")

	// Extract environment variable for 'ThisDomainsUuid'
	common_config.ThisDomainsUuid = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ThisDomainsUuid")

	// Extract environment variable for 'ThisExecutionDomainUuid'
	common_config.ThisExecutionDomainUuid = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ThisExecutionDomainUuid")

	// Extract if native pubsub client library should be used or not
	common_config.UseNativeGcpPubSubClientLibrary, err = strconv.ParseBool(environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("UseNativeGcpPubSubClientLibrary"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable 'UseNativeGcpPubSubClientLibrary:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract if a New Baseline for TestInstructions, TestInstructionContainers and Users should be saved in database
	common_config.ForceNewBaseLineForTestInstructionsAndTestInstructionContainers, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ForceNewBaseLineForTestInstructionsAndTestInstructionContainers"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'ForceNewBaseLineForTestInstructionsAndTestInstructionContainers:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract if this is the Connector that sends supported TestInstructions, TestInstructionContainers and
	// Users to Worker. If not, then this is only a TestExecutionDomain
	common_config.ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract if PubSubReceiver be started
	common_config.ShouldPubSubReceiverBeStarted, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ShouldPubSubReceiverBeStarted"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'ShouldPubSubReceiverBeStarted:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract if All communication with Worker should be done
	common_config.TurnOffAllCommunicationWithWorker, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("TurnOffAllCommunicationWithWorker"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'TurnOffAllCommunicationWithWorker:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract if Proxy-server should be used for outgoing requests
	common_config.ShouldProxyServerBeUsed, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ShouldProxyServerBeUsed"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'ShouldProxyServerBeUsed:' to an boolean, error: ", err)
		os.Exit(0)
	}

	// Extract URL to Proxy-server for outgoing requests
	common_config.ProxyServerURL = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ProxyServerURL")

	// Extract if SPIRE Server should be used when requesting
	common_config.ShouldSpireServerBeUsedForGettingGcpToken, err = strconv.ParseBool(
		environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ShouldSpireServerBeUsedForGettingGcpToken"))
	if err != nil {
		fmt.Println("Couldn't convert environment variable "+
			"'ShouldSpireServerBeUsedForGettingGcpToken:' to an boolean, error: ", err)
		os.Exit(0)
	}

}
