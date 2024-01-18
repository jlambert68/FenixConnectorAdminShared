package fenixConnectorAdminShared

import (
	"github.com/jlambert68/FenixConnectorAdminShared/connectorEngine"
	"github.com/sirupsen/logrus"
)

type fenixConnectorAdminSharedObjectStruct struct {
	logger                         *logrus.Logger
	TestInstructionExecutionEngine connectorEngine.TestInstructionExecutionEngineStruct
}

// Variable holding everything together
var fenixConnectorAdminSharedObject *fenixConnectorAdminSharedObjectStruct
