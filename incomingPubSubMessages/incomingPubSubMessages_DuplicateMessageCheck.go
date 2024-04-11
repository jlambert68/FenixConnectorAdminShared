package incomingPubSubMessages

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Map used for storing 'UniqueMessageIdentifier'
var (
	uniqueMessageIdentifierMap                    = make(map[string]time.Time)
	checkForDuplicatesAndAddMutex                 sync.Mutex
	checkForProcessedUniqueMessageIdentifierMutex sync.Mutex
	deleteOldUniqueMessageIdentifierMutex         sync.Mutex
	//ctx                                           = context.Background()
	//client                                        *firestore.Client
)

// Function to add a 'UniqueMessageIdentifier' to the local map and Firestore if it not exists
func checkForDuplicatesAndAddForMessages(uniqueMessageIdentifier string) (isDuplicate bool, err error) {

	// Lock for updating the local Map
	checkForDuplicatesAndAddMutex.Lock()
	defer checkForDuplicatesAndAddMutex.Unlock()

	// First check in the local map duplicates
	isDuplicate, _, err = checkForProcessedMessage(uniqueMessageIdentifier)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                      "bf4a210e-42ce-4ba9-a85f-04910feeee7e",
			"uniqueMessageIdentifier": uniqueMessageIdentifier,
			"err":                     err.Error(),
		}).Error("Got some problem when checking for duplicates of processed")

		return false, err
	}

	// Check if this is a duplicate
	if isDuplicate == true {
		return true, err
	}

	// Not a duplicate so add to the local map first
	uniqueMessageIdentifierMap[uniqueMessageIdentifier] = time.Now()

	/*
		// Then add to Firestore
		_, err := client.Collection("uuids").Doc(uuid).Set(ctx, map[string]interface{}{
			"exists": true,
		})
		if err != nil {
			return err
		}
	*/

	// Not a duplicate
	return false, nil
}

// Function to check if 'UniqueMessageIdentifier' exists in First local Map, and if not the in Firestore
func checkForProcessedMessage(
	uniqueMessageIdentifier string) (
	isDuplicate bool,
	addedToMapTimeStamp time.Time,
	err error) {

	// Set a lock for reading the Map if 'checkForDuplicatesAndAddMutex' is ongoing
	if checkForDuplicatesAndAddMutex.TryLock() == true {
		defer checkForDuplicatesAndAddMutex.Unlock()
	}

	// Set a lock for reading the Map if 'deleteOldUniqueMessageIdentifierMutex' is ongoing
	if deleteOldUniqueMessageIdentifierMutex.TryLock() == true {
		defer deleteOldUniqueMessageIdentifierMutex.Unlock()
	}

	// Set a lock for reading the Map if 'checkForDuplicatesAndAddMutex' is ongoing
	checkForProcessedUniqueMessageIdentifierMutex.Lock()
	defer checkForProcessedUniqueMessageIdentifierMutex.Unlock()

	// First check in the local map
	var uniqueMessageIdentifierExistsInMap bool
	addedToMapTimeStamp, uniqueMessageIdentifierExistsInMap = uniqueMessageIdentifierMap[uniqueMessageIdentifier]

	// If the 'uniqueMessageIdentifier' was found in the map then it is a duplicate
	if uniqueMessageIdentifierExistsInMap == true {
		return true, addedToMapTimeStamp, nil // 'uniqueMessageIdentifier' found in local map
	}

	/*
		// Check in Firestore
		_, err := client.Collection("uuids").Doc(uuid).Get(ctx)
		if err != nil {
			if firestore.ErrNotFound == err {
				return false, nil // UUID not found in Firestore
			}
			return false, err // An error occurred
		}
	*/

	// Not a duplicate
	return false, time.Time{}, nil
}

// Function to delete messages older than 1 hour
func deleteOldUniqueMessageIdentifier(uniqueMessageIdentifier string) (err error) {

	// Lock for deleting the local Map
	deleteOldUniqueMessageIdentifierMutex.Lock()
	defer deleteOldUniqueMessageIdentifierMutex.Unlock()

	// Check if it exists and get timestamp for when it was stored
	var addedToMapTimeStamp time.Time
	var isStored bool

	isStored, addedToMapTimeStamp, err = checkForProcessedMessage(uniqueMessageIdentifier)

	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"ID":                      "b7cbc1b0-9f1b-47a0-a7f0-bafcc034495c",
			"uniqueMessageIdentifier": uniqueMessageIdentifier,
			"err":                     err.Error(),
		}).Error("Got some problem when checking for processed messages")
	}

	if isStored == true {

		// Check if timestamp is older than 1 hour, if so then delete it from the map
		if addedToMapTimeStamp.Add(1 * time.Hour).Before(time.Now()) {
			delete(uniqueMessageIdentifierMap, uniqueMessageIdentifier)
		}
	}

	return err
}
