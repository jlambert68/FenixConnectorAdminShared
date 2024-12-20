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
		generateNewPrivatePublicKeyParInGcpSecretManager()

	} else {

		// Only done in GCP
		if common_config.ExecutionLocationForConnector == common_config.GCP &&
			executionEnvironmentPlatform == executionEnvironmentPlatformGCP {

			// Secure that there is a "latest version".
			// One scenario when this is not the case is the first time when deployed
			var secretManagerPath string
			secretManagerPath = fmt.Sprintf(secretManagerPathForPrivateKey,
				common_config.GcpProject) + latestSecretVersion
			_, err = AccessSecretVersion(secretManagerPath)

			if err != nil {
				common_config.Logger.WithFields(logrus.Fields{
					"Id":                "76608804-a09e-4e94-9b4f-ebd075b23479",
					"err":               err,
					"secretManagerPath": secretManagerPath,
				}).Info("No latest version for secrete. Will create a new one")

				// Generates a new Private-Public key par and exit the application for user to be able to update
				// Fenix Database with the new Public key for the Domain
				generateNewPrivatePublicKeyParInGcpSecretManager()

			}
		}

		// Should new Private-Public-key-par be generated (not in GCP)
		if (executionEnvironment == environmentDev && generateNewPrivateKeyForDev == true) ||
			(executionEnvironment == environmentAcc && generateNewPrivateKeyForAcc == true) {

			// Generates a new Private-Public key par and exit the application for user to be able to update
			// Fenix Database with the new Public key for the Domain and set the Private Key as an environment variable
			// for the Connector
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
				"Id":  "723ec3ce-5741-40aa-86ad-5df7e53dbadb",
				"err": err,
			}).Warning("Couldn't generate Public Key from Private Key, Generate a new Private Key and retry to Generate the public key from the new one")

			// If there are any problem generating the public key from the Private key, then create a new private key
			// This can happen in GCP when the environment is first set up
			generateNewPrivatePublicKeyParInGcpSecretManager()

		}

		// Inform of used Public key
		common_config.Logger.WithFields(logrus.Fields{
			"Id":        "45c87663-0ce8-4b0d-81c0-3088131a508a",
			"PublicKey": publicKey,
		}).Info("Will use existing Private Key.")
	}
}

// Generates a new Private-Public key par and exits the application for user to be able to update
// Fenix Database with the new Public key for the Domain
// For GCP
func generateNewPrivatePublicKeyParInGcpSecretManager() {

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
		common_config.GcpProject)

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

// Generates a new Private-Public key par and exits the application for user to be able to update
// Fenix Database with the new Public key for the Domain
func generateNewPrivatePublicKeyPar() {

	var err error

	// Generate Private Key
	var privateKey string
	privateKey, err = shared_code.GenerateNewPrivateKeyAsBase64String()
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":  "8a38e3cb-afaa-4bca-82a2-e0da91349f07",
			"err": err,
		}).Fatalln("Couldn't generate Private Key, so will exit")
	}

	// Generate Public Key from private key
	var publicKey string
	publicKey, err = shared_code.GeneratePublicKeyAsBase64StringFromPrivateKeyInput(privateKey)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":  "c63d41df-12ff-4ba7-ac0c-e7cce53fd378",
			"err": err,
		}).Fatalln("Couldn't generate Public Key from Private Key, so will exit")
	}

	// Inform that new PrivateKey was successfully generated
	common_config.Logger.WithFields(logrus.Fields{
		"Id":         "9498f7b4-6216-40e7-a845-abef00d802fa",
		"PublicKey":  publicKey,
		"PrivateKey": privateKey,
	}).Info("Successfully generated and new PrivateKey-PublicKey-par.")

	// Exiting so public key can be transferred to Fenix Database
	os.Exit(0)

}
