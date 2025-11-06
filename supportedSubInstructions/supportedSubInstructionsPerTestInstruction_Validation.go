package supportedSubInstructions

import (
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sirupsen/logrus"
)

//
//go:embed json-schema/supportedSubInstructionsPerTestInstruction.json-schema.json
var embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema []byte

// ValidateSupportedSubInstructionsPerTestInstruction
// Validates the SupportedSubInstructionsPerTestInstruction
func ValidateSupportedSubInstructionsPerTestInstruction(
	supportedSubInstructionsJsonTAsByteArrayPtr *[]byte,
	supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr *[][]byte) (err error) {

	// Validates the SupportedSubInstructionsPerTestInstruction-json towards the json-schema
	err = validateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema(
		supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr)
	if err != nil {
		return err
	}

	// Validates that the ExecutionOrder towards the PreConditions for SupportedSubInstructionsPerTestInstruction
	err = validateExecutionOrderTowardsPreConditionsForSupportedSubInstructionsPerTestInstructionJson(
		supportedSubInstructionsJsonTAsByteArrayPtr,
		supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr)
	if err != nil {
		return err
	}

	return err
}

// ValidateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema
// Validates the SupportedSubInstructionsPerTestInstruction-json towards the json-schema
func validateSupportedSubInstructionsPerTestInstructionJsonTowardsJsonSchema(
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

	return err

}

// Validates that the ExecutionOrder towards the PreConditions for SupportedSubInstructionsPerTestInstruction
func validateExecutionOrderTowardsPreConditionsForSupportedSubInstructionsPerTestInstructionJson(
	supportedSubInstructionsJsonTAsByteArrayPtr *[]byte,
	supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr *[][]byte) (
	err error) {

	var supportedSubInstructionsAsCatalog Catalog
	err = json.Unmarshal(*supportedSubInstructionsJsonTAsByteArrayPtr, &supportedSubInstructionsAsCatalog)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "d80e6388-3977-4dd3-88f1-f038034d488b",
			"err": err,
			"string(*supportedSubInstructionsJsonTAsByteArrayPtr)": string(*supportedSubInstructionsJsonTAsByteArrayPtr),
		}).Error("Couldn't UnMarshal 'supportedSubInstructionsJsonTAsByteArrayPtr' into struct used for processing ExecutionPlan")

		return err
	}

	// UnMarshal json-schema for SupportedSubInstructionsPerTestInstruction
	var supportedSubInstructionsPerTestInstructionJsonSchema PlanRoot
	if err = json.Unmarshal(embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema, &supportedSubInstructionsPerTestInstructionJsonSchema); err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "3ad2ad8c-4ce3-4a77-8240-62785d25c869",
			"err": err,
			"string(embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema)": string(embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema),
		}).Error("Couldn't UnMarshal 'embeddedFile_SupportedSubInstructionsPerTestInstructionJsonSchema' into struct used for processing ExecutionPlan")

		return err
	}

	// Convert to object that can be validated
	var supportedSubInstructionsPerTestInstructionToValidateAsByteArray [][]byte
	supportedSubInstructionsPerTestInstructionToValidateAsByteArray = *supportedSubInstructionsPerTestInstructionToValidateAsByteArrayPtr

	// Loop all SupportedSubInstructionsPerTestInstruction To be Validated
	for _, tempSupportedSubInstructionsPerTestInstructionToValidated := range supportedSubInstructionsPerTestInstructionToValidateAsByteArray {

		// Validate that the 'supportedSubInstructionsJson' is valid towards the json-schema
		// UnMarshal
		var supportedSubInstructionsPerTestInstructionToValidatedAsPlanRoot PlanRoot
		if err = json.Unmarshal(tempSupportedSubInstructionsPerTestInstructionToValidated, &supportedSubInstructionsPerTestInstructionToValidatedAsPlanRoot); err != nil {
			common_config.Logger.WithFields(logrus.Fields{
				"id":  "f95b0c2e-a0d1-437d-95ba-ebd2946821b9",
				"err": err,
				"tempSupportedSubInstructionsPerTestInstructionToValidated": string(tempSupportedSubInstructionsPerTestInstructionToValidated),
			}).Error("Couldn't UnMarshal 'tempSupportedSubInstructionsPerTestInstructionToValidated' into struct used for processing ExecutionPlan")

			return err
		}

		// --- Validate supportedSubInstructionsPerTestInstructionToValidatedAsPlanRoot ---
		result := ValidatePlan(supportedSubInstructionsAsCatalog, supportedSubInstructionsPerTestInstructionToValidatedAsPlanRoot)

		// --- Print results ---
		if !result.OK {

		}
		common_config.Logger.WithFields(logrus.Fields{
			"id":  "8e6454cd-bb7c-4252-897f-e7291048c037",
			"err": result.Errors,
			"tempSupportedSubInstructionsPerTestInstructionToValidated": string(tempSupportedSubInstructionsPerTestInstructionToValidated),
		}).Error("Failed when validating ExecutionOrder")

		err = errors.New("failed when validating PreConditions")

		return err

		return err

	}

	return err

}
