package incomingPubSubMessages

import "github.com/jlambert68/FenixConnectorAdminShared/common_config"

// Create the PubSub-topic from Domain-Uuid
func generatePubSubTopicNameForExecutionStatusUpdates() (statusExecutionTopic string) {

	var pubSubTopicBase string
	pubSubTopicBase = common_config.TestInstructionExecutionPubSubTopicBase

	// Get the first 8 characters from ThisDomainsUuid
	var shortedThisDomainsUuid string
	shortedThisDomainsUuid = common_config.ThisDomainsUuid[0:8]

	// Get the first 8 characters from 'thisExecutionDomainUuid'
	var shortedThisExecutionDomainUuid string
	shortedThisExecutionDomainUuid = common_config.ThisExecutionDomainUuid[0:8]

	// Build PubSub-topic
	statusExecutionTopic = pubSubTopicBase + "-" + shortedThisDomainsUuid + "-" + shortedThisExecutionDomainUuid

	return statusExecutionTopic
}

// Creates a Topic-Subscription-Name
func generatePubSubTopicSubscriptionNameForExecutionStatusUpdates() (topicSubscriptionName string) {

	const topicSubscriptionPostfix string = "-sub"

	// Get Topic-name
	var topicID string
	topicID = generatePubSubTopicNameForExecutionStatusUpdates()

	// Create the Topic-Subscription-name
	topicSubscriptionName = topicID + topicSubscriptionPostfix

	return topicSubscriptionName
}
