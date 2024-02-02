package gRPCServer

import (
	fenixExecutionConnectorGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionConnectorGrpcApi/go_grpc_api"
	"google.golang.org/grpc"
	"net"
)

// gRPCServer variables
var (
	fenixExecutionConnectorGrpcServer *grpc.Server
	//registerFenixExecutionConnectorGrpcServicesServer       *grpc.Server
	//registerFenixExecutionConnectorWorkerGrpcServicesServer *grpc.Server
	lis net.Listener
)

// The object for the 'FenixConnectorGrpcServicesServerStruct'
var FenixConnectorGrpcServicesServerObject *FenixConnectorGrpcServicesServerStruct

// gRPCServer Server type
type FenixConnectorGrpcServicesServerStruct struct {
	fenixExecutionConnectorGrpcApi.UnimplementedFenixExecutionConnectorGrpcServicesServer
}
