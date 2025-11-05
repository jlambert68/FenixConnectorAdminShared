package supportedSubInstructions

import (
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//go:embed json-schema/supportedSubInstructions.json-schema.json
var embeddedFile_SupportedSubInstructionsJsonSchema []byte

// ValidateSupportedSubInstructions
// Validates the SupportedSubInstructions-json
func ValidateSupportedSubInstructions(
	supportedSubInstructionsJsonToValidateAsByteArrayPtr *[]byte) (
	err error) {

	// Validates the SupportedSubInstructions-json towards the json-schema
	err = validateSupportedSubInstructionsJsonTowardsJsonSchema(supportedSubInstructionsJsonToValidateAsByteArrayPtr)
	if err != nil {
		return err
	}

	// Validates that the PreConditions aren't broken for SupportedSubInstructions-json
	err = validatePreConditionsForSupportedSubInstructionsJson(supportedSubInstructionsJsonToValidateAsByteArrayPtr)
	if err != nil {
		return err
	}

	return err
}

// ValidatePreConditionsForSupportedSubInstructionsJson
// Validates that the PreConditions aren't broken for SupportedSubInstructions-json
func validatePreConditionsForSupportedSubInstructionsJson(supportedSubInstructionsJsonToValidateAsByteArrayPtr *[]byte) (err error) {

	var doc SupportedSubInstructionsDocument
	if err = json.Unmarshal(*supportedSubInstructionsJsonToValidateAsByteArrayPtr, &doc); err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "43ba097e-6e5b-4a7c-8c2f-a746acd3928c",
			"err": err,
			"string(supportedSubInstructionsJsonToValidateAsByteArrayPtr)": string(*supportedSubInstructionsJsonToValidateAsByteArrayPtr),
		}).Error("Couldn't UnMarshal 'supportedSubInstructionsJsonToValidateAsByteArrayPtr' into struct used for processing PreConditions")

		return err
	}

	res := ValidatePreconditions(doc, true)
	if !res.OK {

		common_config.Logger.WithFields(logrus.Fields{
			"id":  "25c1457d-6a6b-420d-9e4a-310a7e42fe00",
			"err": res.Errors,
			"string(supportedSubInstructionsJsonToValidateAsByteArrayPtr)": string(*supportedSubInstructionsJsonToValidateAsByteArrayPtr),
		}).Error("Failed when validating PreConditions")

		err = errors.New("failed when validating PreConditions")

		return err

	}

	return err
}

// ValidateSupportedSubInstructionsJsonTowardsJsonSchema
// Validates the SupportedSubInstructions-json towards the json-schema
func validateSupportedSubInstructionsJsonTowardsJsonSchema(
	supportedSubInstructionsJsonToValidateAsByteArrayPtr *[]byte) (
	err error) {

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
