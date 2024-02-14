package common_config

// ***********************************************************************************************************
// The following variables receives their values from environment variables

// ExecutionLocationForConnector - Where is the Worker running
var ExecutionLocationForConnector ExecutionLocationTypeType

// ExecutionLocationForFenixExecutionWorkerServer  - Where is Fenix Execution Server running
var ExecutionLocationForFenixExecutionWorkerServer ExecutionLocationTypeType

// ExecutionLocationTypeType - Definitions for where client and Fenix Server is running
type ExecutionLocationTypeType int

// Constants used for where stuff is running
const (
	LocalhostNoDocker ExecutionLocationTypeType = iota
	LocalhostDocker
	GCP
)

// Address to Fenix Execution Worker & Execution Connector, will have their values from Environment variables at startup
var (
	FenixExecutionWorkerAddress                                                           string
	FenixExecutionWorkerPort                                                              int
	FenixExecutionWorkerAddressToDial                                                     string
	ExecutionConnectorPort                                                                int
	GCPAuthentication                                                                     bool
	AuthClientId                                                                          string
	AuthClientSecret                                                                      string
	UseServiceAccount                                                                     bool
	GcpProject                                                                            string
	UsePubSubToReceiveMessagesFromWorker                                                  bool
	ShouldPubSubReceiverBeStarted                                                         bool
	LocalServiceAccountPath                                                               string
	TestInstructionExecutionPubSubTopicBase                                               string
	ThisDomainsUuid                                                                       string
	ThisExecutionDomainUuid                                                               string
	UseNativeGcpPubSubClientLibrary                                                       bool
	ForceNewBaseLineForTestInstructionsAndTestInstructionContainers                       bool
	ThisConnectorIsTheOneThatPublishSupportedTestInstructionsAndTestInstructionContainers bool
	TurnOffAllCommunicationWithWorker                                                     bool
)
