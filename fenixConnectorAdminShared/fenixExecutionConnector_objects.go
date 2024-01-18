package fenixConnectorAdminShared

import (
	"github.com/sirupsen/logrus"
)

type fenixConnectorAdminSharedObjectStruct struct {
	logger *logrus.Logger
}

// Variable holding everything together
var fenixConnectorAdminSharedObject *fenixConnectorAdminSharedObjectStruct
