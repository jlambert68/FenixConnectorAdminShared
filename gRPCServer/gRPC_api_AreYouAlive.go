package gRPCServer

import (
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"
	fenixExecutionConnectorGrpcApi "github.com/jlambert68/FenixGrpcApi/FenixExecutionServer/fenixExecutionConnectorGrpcApi/go_grpc_api"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"time"
)

// AreYouAlive - *********************************************************************
// Anyone can check if Execution Connector server is alive with this service
func (fenixConnectorGrpcObject *FenixConnectorGrpcServicesServerStruct) AreYouAlive(
	ctx context.Context,
	emptyParameter *fenixExecutionConnectorGrpcApi.EmptyParameter) (
	*fenixExecutionConnectorGrpcApi.AckNackResponse, error) {

	common_config.Logger.WithFields(logrus.Fields{
		"id": "1ff67695-9a8b-4821-811d-0ab8d33c4d8b",
	}).Debug("Incoming 'gRPCServer - AreYouAlive'")

	common_config.Logger.WithFields(logrus.Fields{
		"id": "9c7f0c3d-7e9f-4c91-934e-8d7a22926d84",
	}).Debug("Outgoing 'gRPCServer - AreYouAlive'")

	ackNackResponse := &fenixExecutionConnectorGrpcApi.AckNackResponse{
		AckNack:                         true,
		Comments:                        fmt.Sprintf("I'am alive and the time is %s", time.Now().String()),
		ErrorCodes:                      nil,
		ProtoFileVersionUsedByConnector: fenixExecutionConnectorGrpcApi.CurrentFenixExecutionConnectorProtoFileVersionEnum(common_config.GetHighestConnectorProtoFileVersion()),
	}

	return ackNackResponse, nil
}
