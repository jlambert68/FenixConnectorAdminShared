package messagesToExecutionWorkerServer

import (
	"github.com/jlambert68/FenixConnectorAdminShared/gcp"
	fenixExecutionWorkerGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionWorkerGrpcApi/go_grpc_api"
	"google.golang.org/grpc"
)

type MessagesToExecutionWorkerObjectStruct struct {
	Gcp                         gcp.GcpObjectStruct
	connectionToWorkerInitiated bool
}

// Variables used for contacting Fenix Execution Worker Server
var (
	remoteFenixExecutionWorkerServerConnection *grpc.ClientConn
	fenixExecutionWorkerGrpcClient             fenixExecutionWorkerGrpcApi.FenixExecutionWorkerConnectorGrpcServicesClient
)
