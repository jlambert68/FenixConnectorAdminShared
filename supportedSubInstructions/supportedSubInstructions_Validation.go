package supportedSubInstructions

import (
	_ "embed"
	"encoding/json"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//go:embed json-schema/supportedSubInstructions.json-schema.json
var embeddedFile_SupportedSubInstructionsJsonSchema []byte

// ValidateSupportedSubInstructionsJsonTowardsJsonSchema
// Validates the SupportedMetaData-json towards the json-schema
func ValidateSupportedSubInstructionsJsonTowardsJsonSchema(
	supportedSubInstructionsJsonToValidateAsByteArrayPtr *[]byte) (err error) {

	// Get the json-schema as string
	var supportedSubInstructionsJsonSchemaAsString string
	supportedSubInstructionsJsonSchemaAsString = string(embeddedFile_SupportedSubInstructionsJsonSchema)

	// Compile json-schema 'supportedSubInstructionsJsonSchema'
	var supportedSubInstructionsJsonSchema *jsonschema.Schema
	supportedSubInstructionsJsonSchema, err = jsonschema.CompileString("schema.json", supportedSubInstructionsJsonSchemaAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "450ebcf0-b716-4b59-bb21-0f64f4344083",
			"err": err,
			"supportedSubInstructionsJsonSchemaAsString": supportedSubInstructionsJsonSchemaAsString,
		}).Error("Couldn't compile the json-schema for 'supportedSubInstructionsJsonSchemaAsString'")

		return err

	}

	// Convert to object that can be validated
	var supportedSubInstructionsJsonToValidateAsByteArray []byte
	supportedSubInstructionsJsonToValidateAsByteArray = *supportedSubInstructionsJsonToValidateAsByteArrayPtr
	var jsonObjectedToBeValidated interface{}
	err = json.Unmarshal(supportedSubInstructionsJsonToValidateAsByteArray, &jsonObjectedToBeValidated)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "2832cf42-52aa-4baa-8da1-77259e262d60",
			"err": err,
			"string(supportedSubInstructionsJsonToValidateAsByteArray)": string(supportedSubInstructionsJsonToValidateAsByteArray),
		}).Error("Couldn't Unmarshal 'supportedSubInstructionsJsonToValidateAsByteArray'")

		return err
	}

	// Validate that the 'supportedSubInstructionsJson' is valid towards the json-schema

	err = supportedSubInstructionsJsonSchema.Validate(jsonObjectedToBeValidated)
	if err != nil {

		// json is not valid towards json-schema
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "97afd5e5-2e12-4ec7-b0f8-e988c07892c7",
			"err": err,
			"string(supportedSubInstructionsJsonToValidateAsByteArray)": string(supportedSubInstructionsJsonToValidateAsByteArray),
			"supportedSubInstructionsJsonSchemaAsString":                supportedSubInstructionsJsonSchemaAsString,
		}).Error("'supportedMetaDataJsonToValidateAsString' is not valid to json-schema " +
			"'supportedSubInstructionsJsonSchemaAsString'")

		return err

	} else {
		// json is valid towards json-schema
		/*
			common_config.Logger.WithFields(logrus.Fields{
				"id": "a61ad3b1-63db-4a0a-b39f-0526d52fbcde",
				"string(supportedSubInstructionsJsonToValidateAsByteArray)": string(supportedSubInstructionsJsonToValidateAsByteArray),
				"supportedSubInstructionsJsonSchemaAsString":                supportedSubInstructionsJsonSchemaAsString,
			}).Debug("'supportedMetaDataJsonToValidateAsString' is valid to json-schema " +
				"'supportedSubInstructionsJsonSchemaAsString'")

		*/
	}

	return nil

}
