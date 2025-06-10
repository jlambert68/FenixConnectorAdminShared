package supportedMetaData

import (
	_ "embed"
	"encoding/json"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//go:embed json-schema/supportedTestCaseMetaData.json-schema.json
var embeddedFile_SupportedTestCaseMetaDataJsonSchema []byte

// ValidateSupportedTestCaseMetaDataJsonTowardsJsonSchema
// Validates the SupportedMetaData-json towards the json-schema
func ValidateSupportedTestCaseMetaDataJsonTowardsJsonSchema(
	supportedMetaDataJsonToValidateAsByteArrayPtr *[]byte) (err error) {

	// Get the json-schema as string
	var supportedTestCaseMetaDataJsonSchemaAsString string
	supportedTestCaseMetaDataJsonSchemaAsString = string(embeddedFile_SupportedTestCaseMetaDataJsonSchema)

	// Compile json-schema 'supportedTestCaseMetaDataJsonSchema'
	var supportedTestCaseMetaDataJsonSchema *jsonschema.Schema
	supportedTestCaseMetaDataJsonSchema, err = jsonschema.CompileString("schema.json", supportedTestCaseMetaDataJsonSchemaAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "b26241f4-34aa-4b81-991c-25bb523800ee",
			"err": err,
			"supportedTestCaseMetaDataJsonSchemaAsString": supportedTestCaseMetaDataJsonSchemaAsString,
		}).Error("Couldn't compile the json-schema for 'supportedTestCaseMetaDataJsonSchemaAsString'")

		return err

	}

	// Convert to object that can be validated
	var supportedMetaDataJsonToValidateAsByteArray []byte
	supportedMetaDataJsonToValidateAsByteArray = *supportedMetaDataJsonToValidateAsByteArrayPtr
	var jsonObjectedToBeValidated interface{}
	err = json.Unmarshal(supportedMetaDataJsonToValidateAsByteArray, &jsonObjectedToBeValidated)

	// Validate that the 'supportedMetaDataJson' is valid towards the json-schema

	err = supportedTestCaseMetaDataJsonSchema.Validate(jsonObjectedToBeValidated)
	if err != nil {

		// json is not valid towards json-schema
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "a35cb2e5-ba30-4a01-9c09-0676da84f65f",
			"err": err,
			"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
			"supportedTestCaseMetaDataJsonSchemaAsString":        supportedTestCaseMetaDataJsonSchemaAsString,
		}).Error("'supportedMetaDataJsonToValidateAsString' is not valid to json-schema " +
			"'supportedTestCaseMetaDataJsonSchemaAsString'")

		return err

	} else {
		// json is valid towards json-schema
		/*
			common_config.Logger.WithFields(logrus.Fields{
				"id": "a61ad3b1-63db-4a0a-b39f-0526d52fbcde",
				"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
				"supportedTestCaseMetaDataJsonSchemaAsString":                supportedTestCaseMetaDataJsonSchemaAsString,
			}).Debug("'supportedMetaDataJsonToValidateAsString' is valid to json-schema " +
				"'supportedTestCaseMetaDataJsonSchemaAsString'")

		*/
	}

	return nil

}
