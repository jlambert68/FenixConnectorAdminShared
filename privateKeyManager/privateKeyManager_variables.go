package privateKeyManager

// Definitions for in what executionEnvironment Cloud Run Service is running
type executionEnvironmentTypeType uint

const (
	environmentDev executionEnvironmentTypeType = iota
	environmentAcc
)

// Definitions for in what executionEnvironmentPlatform
type executionEnvironmentPlatformTypeType uint

const (
	executionEnvironmentPlatformGCP executionEnvironmentPlatformTypeType = iota
	executionEnvironmentPlatformSebShift
	executionEnvironmentPlatformOther
)

// *** START Environment variables START ***

// Should a new Private-Public-key-par be generated for System Environment
var generateNewPrivateKeyForDev bool

// Should a new Private-Public-key-par be generated for Acceptance Environment
var generateNewPrivateKeyForAcc bool

// Environment (dev or acc)
var executionEnvironment executionEnvironmentTypeType
var executionEnvironmentAsString string

// ExecutionEnvironmentPlatform (GCP, SebShift, Other)
var executionEnvironmentPlatform executionEnvironmentPlatformTypeType
var executionEnvironmentPlatformAsString string

// *** END Environment variables END ***

// Path to the latest secret
const latestSecretVersion = "/versions/latest"

// Path to the first secret
const firstSecretVersion = "/versions/1"

const secretManagerPathForPrivateKey = "projects/%s-%s/secrets/private-key"
