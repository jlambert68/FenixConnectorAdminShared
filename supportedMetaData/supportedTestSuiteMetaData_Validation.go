package supportedMetaData

import (
	_ "embed"
	"encoding/json"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//go:embed json-schema/supportedTestSuiteMetaData.json-schema.json
var embeddedFile_SupportedTestSuiteMetaDataJsonSchema []byte

// ValidateSupportedTestSuiteMetaDataJsonTowardsJsonSchema
// Validates the SupportedMetaData-json towards the json-schema
func ValidateSupportedTestSuiteMetaDataJsonTowardsJsonSchema(
	supportedMetaDataJsonToValidateAsByteArrayPtr *[]byte) (err error) {

	// Get the json-schema as string
	var supportedTestSuiteMetaDataJsonSchemaAsString string
	supportedTestSuiteMetaDataJsonSchemaAsString = string(embeddedFile_SupportedTestSuiteMetaDataJsonSchema)

	// Compile json-schema 'supportedTestSuiteMetaDataJsonSchema'
	var supportedTestSuiteMetaDataJsonSchema *jsonschema.Schema
	supportedTestSuiteMetaDataJsonSchema, err = jsonschema.CompileString("schema.json", supportedTestSuiteMetaDataJsonSchemaAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "067d8971-f445-4258-8d55-996accb044c1",
			"err": err,
			"supportedTestSuiteMetaDataJsonSchemaAsString": supportedTestSuiteMetaDataJsonSchemaAsString,
		}).Error("Couldn't compile the json-schema for 'supportedTestSuiteMetaDataJsonSchemaAsString'")

		return err

	}

	// Convert to object that can be validated
	var supportedMetaDataJsonToValidateAsByteArray []byte
	supportedMetaDataJsonToValidateAsByteArray = *supportedMetaDataJsonToValidateAsByteArrayPtr
	var jsonObjectedToBeValidated interface{}
	err = json.Unmarshal(supportedMetaDataJsonToValidateAsByteArray, &jsonObjectedToBeValidated)

	// Validate that the 'supportedMetaDataJson' is valid towards the json-schema

	err = supportedTestSuiteMetaDataJsonSchema.Validate(jsonObjectedToBeValidated)
	if err != nil {

		// json is not valid towards json-schema
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "4f4a2827-1f57-4475-a0ca-889016fa5a5a",
			"err": err,
			"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
			"supportedTestSuiteMetaDataJsonSchemaAsString":       supportedTestSuiteMetaDataJsonSchemaAsString,
		}).Error("'supportedMetaDataJsonToValidateAsString' is not valid to json-schema " +
			"'supportedTestSuiteMetaDataJsonSchemaAsString'")

		return err

	} else {
		// json is valid towards json-schema
		/*
			common_config.Logger.WithFields(logrus.Fields{
				"id": "3589a954-1d8b-45a2-94fe-eca1d2f79fc9",
				"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
				"supportedTestSuiteMetaDataJsonSchemaAsString":                supportedTestSuiteMetaDataJsonSchemaAsString,
			}).Debug("'supportedMetaDataJsonToValidateAsString' is valid to json-schema " +
				"'supportedTestSuiteMetaDataJsonSchemaAsString'")

		*/
	}

	return nil

}
