package privateKeyManager

import (
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/jlambert68/FenixSyncShared/environmentVariables"
	"github.com/jlambert68/FenixTestInstructionsAdminShared/shared_code"
	"github.com/sirupsen/logrus"
	"os"
)

func ShouldNewPrivateKeyBeGenerated() {

	var err error

	// Only create a new PrivateKey if program is running in GCP and in correct executionEnvironment and
	// if parameter for that specific executionEnvironment say that a new private key should be generated
	if common_config.ExecutionLocationForConnector == common_config.GCP &&
		executionEnvironmentPlatform == executionEnvironmentPlatformGCP &&
		((executionEnvironment == environmentDev && generateNewPrivateKeyForDev == true) ||
			(executionEnvironment == environmentAcc && generateNewPrivateKeyForAcc == true)) {

		// Generates a new Private-Public key par and exits the application for user to be able to update
		// Fenix Database with the new Public key for the Domain
		generateNewPrivatePublicKeyPar()

	} else {

		// Secure that there is a "latest version".
		// One scenario when this is not the case is the first time when deployed
		var secretManagerPath string
		secretManagerPath = fmt.Sprintf(secretManagerPathForPrivateKey,
			common_config.GcpProject, executionEnvironmentAsString)
		_, err = AccessSecretVersion(secretManagerPath)

		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"Id":                "76608804-a09e-4e94-9b4f-ebd075b23479",
				"err":               err,
				"secretManagerPath": secretManagerPath,
			}).Info("No latest version for secrete. Will create a new one")

			// Generates a new Private-Public key par and exits the application for user to be able to update
			// Fenix Database with the new Public key for the Domain
			generateNewPrivatePublicKeyPar()

		}

		// Get Public Key from existing private key
		var privateKeyFromEnvironmentVariables string
		privateKeyFromEnvironmentVariables = environmentVariables.
			ExtractEnvironmentVariableOrInjectedEnvironmentVariable("PrivateKey")

		var publicKey string
		publicKey, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKeyInput(privateKeyFromEnvironmentVariables)
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"Id":  "8d332de0-dfe9-48b5-80b7-c87db6307157",
				"err": err,
			}).Fatalln("Couldn't generate Public Key from Private Key, so will exit")
		}

		// Inform of used Public key
		common_config.Logger.WithFields(logrus.Fields{
			"Id":        "45c87663-0ce8-4b0d-81c0-3088131a508a",
			"PublicKey": publicKey,
		}).Info("Will use existing Private Key in in Secret Manager.")
	}
}

// Generates a new Private-Public key par and exits the application for user to be able to update
// Fenix Database with the new Public key for the Domain
func generateNewPrivatePublicKeyPar() {

	var err error

	// Generate Private Key
	var privateKey string
	privateKey, err = shared_code.GenerateNewPrivateKeyAsBase64String()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":  "af764075-9bac-4ff1-bbb8-3a4fc2c73468",
			"err": err,
		}).Fatalln("Couldn't generate Private Key, so will exit")
	}

	// Generate Public Key from private key
	var publicKey string
	publicKey, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKeyInput(privateKey)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":  "43524a4d-1809-4501-a971-571288f2a38d",
			"err": err,
		}).Fatalln("Couldn't generate Public Key from Private Key, so will exit")
	}

	// Store New Private Key in Secret Manager
	var secretManagerPath string
	secretManagerPath = fmt.Sprintf(secretManagerPathForPrivateKey,
		common_config.GcpProject, executionEnvironmentAsString)

	_, err = AddSecretVersion(secretManagerPath, privateKey)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":                "a4667ed3-c109-4f12-8cf8-c6447dfb2f0c",
			"err":               err,
			"secretManagerPath": secretManagerPath,
		}).Fatalln("Couldn't store Private Key in Secret Manager, so will exit")
	}

	// Store the private key in the environment variable 'PrivateKey'
	err = os.Setenv("PrivateKey", privateKey)
	if err != nil {
		defer common_config.Logger.WithFields(logrus.Fields{
			"Id":  "09c65ab9-26cb-4966-a09a-71e706499310",
			"err": err,
		}).Fatalln("Couldn't set value for environment variable PrivateKey, so will exit")
	}

	// Inform that new PrivateKey was successfully generated
	common_config.Logger.WithFields(logrus.Fields{
		"Id":        "b277b744-0fa4-4438-aea9-e47057c298a6",
		"PublicKey": publicKey,
	}).Info("Successfully generated and stored new PrivateKey in Secret Manager.")

	// Exiting so public key can be transferred to Fenix Database
	os.Exit(0)

}
