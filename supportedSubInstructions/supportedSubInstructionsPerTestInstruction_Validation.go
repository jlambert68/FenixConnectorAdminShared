package supportedSubInstructions

import (
	_ "embed"
	"encoding/json"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//
//go:embed json-schema/supportedSubInstructionsPerTestInstruction.json-schema.json
var embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema []byte

// ValidateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema
// Validates the SupportedMetaData-json towards the json-schema
func ValidateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema(
	supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr *[][]byte) (err error) {

	// Get the json-schema as string
	var supportedSubInstructionsPerTestInstructionJsonSchemaAsString string
	supportedSubInstructionsPerTestInstructionJsonSchemaAsString = string(embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema)

	// Compile json-schema 'supportedSubInstructionsPerTestInstructionJsonSchema'
	var supportedSubInstructionsPerTestInstructionJsonSchema *jsonschema.Schema
	supportedSubInstructionsPerTestInstructionJsonSchema, err = jsonschema.CompileString("schema.json",
		supportedSubInstructionsPerTestInstructionJsonSchemaAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "8bce35b3-82a5-480b-8673-fcc7b2bdcdf1",
			"err": err,
			"supportedSubInstructionsPerTestInstructionJsonSchemaAsString": supportedSubInstructionsPerTestInstructionJsonSchemaAsString,
		}).Error("Couldn't compile the json-schema for 'supportedSubInstructionsPerTestInstructionJsonSchemaAsString'")

		return err

	}

	// Convert to object that can be validated
	var supportedSubInstructionsPerTestInstructionToValidateAsByteArray [][]byte
	supportedSubInstructionsPerTestInstructionToValidateAsByteArray = *supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr
	var jsonObjectedToBeValidated interface{}

	// Loop all SupportedSubInstructionsPerTestInstruction To be Validated
	for _, tempSupportedSubInstructionsPerTestInstructionToValidated := range supportedSubInstructionsPerTestInstructionToValidateAsByteArray {

		err = json.Unmarshal(tempSupportedSubInstructionsPerTestInstructionToValidated, &jsonObjectedToBeValidated)
		if err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"id":  "fff02bd5-3599-4083-8e7a-00af981800d5",
				"err": err,
				"string(tempSupportedSubInstructionsPerTestInstructionToValidated)": string(tempSupportedSubInstructionsPerTestInstructionToValidated),
			}).Error("Couldn't Unmarshal 'tempSupportedSubInstructionsPerTestInstructionToValidated'")

			return err
		}

		// Validate that the 'supportedSubInstructionsJson' is valid towards the json-schema

		err = supportedSubInstructionsPerTestInstructionJsonSchema.Validate(jsonObjectedToBeValidated)
		if err != nil {

			// json is not valid towards json-schema
			common_config.Logger.WithFields(logrus.Fields{
				"id":  "8d6459fa-af56-4db2-98af-2117a5553a21",
				"err": err,
				"string(tempSupportedSubInstructionsPerTestInstructionToValidated)": string(tempSupportedSubInstructionsPerTestInstructionToValidated),
				"supportedSubInstructionsPerTestInstructionJsonSchemaAsString":      supportedSubInstructionsPerTestInstructionJsonSchemaAsString,
			}).Error("'tempSupportedSubInstructionsPerTestInstructionToValidated' is not valid to json-schema " +
				"'supportedSubInstructionsPerTestInstructionJsonSchemaAsString'")

			return err

		} else {
			// json is valid towards json-schema
			/*
				common_config.Logger.WithFields(logrus.Fields{
					"id": "a61ad3b1-63db-4a0a-b39f-0526d52fbcde",
					"string(supportedSubInstructionsPerTestInstructionToValidateAsByteArray)": string(supportedSubInstructionsPerTestInstructionToValidateAsByteArray),
					"supportedSubInstructionsJsonSchemaAsString":                supportedSubInstructionsJsonSchemaAsString,
				}).Debug("'supportedMetaDataJsonToValidateAsString' is valid to json-schema " +
					"'supportedSubInstructionsJsonSchemaAsString'")

			*/
		}

	}

	return nil

}
