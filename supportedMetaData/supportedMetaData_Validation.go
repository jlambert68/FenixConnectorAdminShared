package supportedMetaData

import (
	_ "embed"
	"encoding/json"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//go:embed json-schema/supportedMetaData.json-schema.json
var embeddedFile_SupportedMetaDataJsonSchema []byte

// ValidateSupportedMetaDataJsonTowardsJsonSchema
// Validates the SupportedMetaData-json towards the json-schema
func ValidateSupportedMetaDataJsonTowardsJsonSchema(
	supportedMetaDataJsonToValidateAsByteArrayPtr *[]byte) (err error) {

	// Get the json-schema as string
	var supportedMetaDataJsonSchemaAsString string
	supportedMetaDataJsonSchemaAsString = string(embeddedFile_SupportedMetaDataJsonSchema)

	// Compile json-schema 'supportedMetaDataJsonSchema'
	var supportedMetaDataJsonSchema *jsonschema.Schema
	supportedMetaDataJsonSchema, err = jsonschema.CompileString("schema.json", supportedMetaDataJsonSchemaAsString)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":                                  "b26241f4-34aa-4b81-991c-25bb523800ee",
			"err":                                 err,
			"supportedMetaDataJsonSchemaAsString": supportedMetaDataJsonSchemaAsString,
		}).Error("Couldn't compile the json-schema for 'supportedMetaDataJsonSchemaAsString'")

		return err

	}

	// Convert to object that can be validated
	var supportedMetaDataJsonToValidateAsByteArray []byte
	supportedMetaDataJsonToValidateAsByteArray = *supportedMetaDataJsonToValidateAsByteArrayPtr
	var jsonObjectedToBeValidated interface{}
	err = json.Unmarshal(supportedMetaDataJsonToValidateAsByteArray, &jsonObjectedToBeValidated)

	// Validate that the 'supportedMetaDataJson' is valid towards the json-schema

	err = supportedMetaDataJsonSchema.Validate(jsonObjectedToBeValidated)
	if err != nil {

		// json is not valid towards json-schema
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "a35cb2e5-ba30-4a01-9c09-0676da84f65f",
			"err": err,
			"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
			"supportedMetaDataJsonSchemaAsString":                supportedMetaDataJsonSchemaAsString,
		}).Error("'supportedMetaDataJsonToValidateAsString' is not valid to json-schema " +
			"'supportedMetaDataJsonSchemaAsString'")

		return err

	} else {
		// json is valid towards json-schema
		common_config.Logger.WithFields(logrus.Fields{
			"id": "a61ad3b1-63db-4a0a-b39f-0526d52fbcde",
			"string(supportedMetaDataJsonToValidateAsByteArray)": string(supportedMetaDataJsonToValidateAsByteArray),
			"supportedMetaDataJsonSchemaAsString":                supportedMetaDataJsonSchemaAsString,
		}).Debug("'supportedMetaDataJsonToValidateAsString' is valid to json-schema " +
			"'supportedMetaDataJsonSchemaAsString'")
	}

	return nil

}
