package gRPCServer

import (
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	fenixExecutionConnectorGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionConnectorGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"strconv"
)

// InitGrpcServer - Set up and start Backend gRPCServer-server
func (fenixConnectorGrpcObject *FenixConnectorGrpcServicesServerStruct) InitGrpcServer() {

	var err error

	// Find first non allocated port from defined start port
	common_config.Logger.WithFields(logrus.Fields{
		"Id": "054bc0ef-93bb-4b75-8630-74e3823f71da",
	}).Info("Backend Server tries to start")

	common_config.Logger.WithFields(logrus.Fields{
		"Id":                                     "ca3593b1-466b-4536-be91-5e038de178f4",
		"common_config.ExecutionConnectorPort: ": common_config.ExecutionConnectorPort,
	}).Info("Start listening on:")
	lis, err = net.Listen("tcp", ":"+strconv.Itoa(common_config.ExecutionConnectorPort))

	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":    "ad7815b3-63e8-4ab1-9d4a-987d9bd94c76",
			"err: ": err,
		}).Error("failed to listen:")
	} else {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":                                     "ba070b9b-5d57-4c0a-ab4c-a76247a50fd3",
			"common_config.ExecutionConnectorPort: ": common_config.ExecutionConnectorPort,
		}).Info("Success in listening on port:")

	}

	// Create server and register the gRPC-service to the server
	fenixExecutionConnectorGrpcServer = grpc.NewServer()
	fenixExecutionConnectorGrpcApi.RegisterFenixExecutionConnectorGrpcServicesServer(
		fenixExecutionConnectorGrpcServer,
		&FenixConnectorGrpcServicesServerStruct{})

	// Register Reflection on the same server to be able for calling agents to see the methods that are offered
	reflection.Register(fenixExecutionConnectorGrpcServer)

	defer fenixConnectorGrpcObject.StopGrpcServer()

	// Start server
	err = fenixExecutionConnectorGrpcServer.Serve(lis)
	if err != nil {
		common_config.Logger.WithFields(logrus.Fields{
			"Id":    "42abd1b8-2e01-4526-82b4-fb1d6af2b420",
			"err: ": err,
		}).Fatalln("Couldn't start gRPC server")
	}

}

// StopGrpcServer - Stop Backend gRPCServer-server
func (fenixConnectorGrpcObject *FenixConnectorGrpcServicesServerStruct) StopGrpcServer() {

	common_config.Logger.WithFields(logrus.Fields{}).Info("Gracefully stop for: fenixExecutionConnectorGrpcServer")
	fenixExecutionConnectorGrpcServer.GracefulStop()

	common_config.Logger.WithFields(logrus.Fields{
		"common_config.ExecutionConnectorPort: ": common_config.ExecutionConnectorPort,
	}).Info("Closed gRPC-server")

}
