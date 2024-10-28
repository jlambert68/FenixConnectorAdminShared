package privateKeyManager

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixSyncShared/environmentVariables"
	"github.com/sirupsen/logrus"
	"strconv"
)

// Init
// Loads executionEnvironment variables needed by PrivateKey-manager
func Init() {

	var err error
	var tempGenerateNewPrivateKey string

	// Extract executionEnvironment variable 'GenerateNewPrivateKey' for SystemTest (dev)
	tempGenerateNewPrivateKey = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("GenerateNewPrivateKeyForDev")
	generateNewPrivateKeyForDev, err = strconv.ParseBool(tempGenerateNewPrivateKey)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                        "53aedc4c-9687-4be0-84f1-2a5a43fef7b5",
			"error message":             err,
			"tempGenerateNewPrivateKey": tempGenerateNewPrivateKey,
		}).Fatalln("Couldn't parse executionEnvironment variable 'GenerateNewPrivateKeyForDev' as a boolean")
	}

	// Extract executionEnvironment variable 'GenerateNewPrivateKey' for AccTest (acc)
	tempGenerateNewPrivateKey = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("GenerateNewPrivateKeyForAcc")
	generateNewPrivateKeyForAcc, err = strconv.ParseBool(tempGenerateNewPrivateKey)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                        "7bdf00b4-cff6-4597-b3cf-23ee8aa49cff",
			"error message":             err,
			"tempGenerateNewPrivateKey": tempGenerateNewPrivateKey,
		}).Fatalln("Couldn't parse executionEnvironment variable 'GenerateNewPrivateKeyForAcc' as a boolean")
	}

	// Extract executionEnvironment variable 'ENVIRONMENT' (dev or acc)
	var tempEnvironment string
	tempEnvironment = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ENVIRONMENT")

	switch tempEnvironment {

	case "dev":
		executionEnvironment = environmentDev
		executionEnvironmentAsString = tempEnvironment

	case "acc":
		executionEnvironment = environmentAcc
		executionEnvironmentAsString = tempEnvironment

	default:
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                   "88e95aa4-64b1-4037-a277-ffb3b04158d7",
			"executionEnvironment": executionEnvironment,
		}).Fatalln("Environment is not 'dev' or 'acc'")
	}

	// Extract ExecutionEnvironmentPlatform "GCP" or "SebShift" or "Other"
	var tempExecutionEnvironmentPlatform string
	tempExecutionEnvironmentPlatform = environmentVariables.
		ExtractEnvironmentVariableOrInjectedEnvironmentVariable("ExecutionEnvironmentPlatform")

	switch tempExecutionEnvironmentPlatform {

	case "GCP":
		executionEnvironmentPlatform = executionEnvironmentPlatformGCP
		executionEnvironmentPlatformAsString = tempExecutionEnvironmentPlatform

	case "SebShift":
		executionEnvironmentPlatform = executionEnvironmentPlatformSebShift
		executionEnvironmentPlatformAsString = tempExecutionEnvironmentPlatform

	case "Other":
		executionEnvironmentPlatform = executionEnvironmentPlatformOther
		executionEnvironmentPlatformAsString = tempExecutionEnvironmentPlatform

	default:
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                           "dcf406c4-ddc0-4616-abde-822308b9f793",
			"ExecutionEnvironmentPlatform": tempExecutionEnvironmentPlatform,
		}).Fatalln("EnvironmentPlatform is not 'GCP', 'SebShift' or 'Other'")
	}
}
