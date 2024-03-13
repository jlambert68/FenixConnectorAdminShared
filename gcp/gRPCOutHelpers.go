package gcp

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/jlambert68/FenixConnectorAdminShared/common_config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ********************************************************************************************************************

func GRPCDialer(bearerToken string,
	gRpcServerAddress string,
	gRpcServerPort int) (
	*grpc.ClientConn, error) {

	var targetGrpcAddr string
	targetGrpcAddr = gRpcServerAddress + ":" + strconv.Itoa(gRpcServerPort)

	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, // Replace with proper verification
	})

	conn, err := grpc.Dial(
		targetGrpcAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithContextDialer(proxyDialer(common_config.ProxyServerURL, bearerToken)),
		//grpc.WithBlock(),
	)

	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	return conn, err

}

func proxyDialer(proxyURL string, bearerToken string) func(ctx context.Context, addr string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}

		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		// Dial the proxy server
		proxyConn, err := dialer.DialContext(ctx, "tcp", proxy.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to dial proxy server: %w", err)
		}

		// Send an HTTP CONNECT request to the proxy server
		connectReq := &http.Request{
			Method: http.MethodConnect,
			URL:    &url.URL{Opaque: addr},
			Host:   addr,
			Header: make(http.Header),
		}

		// Add the Bearer token to the Authorization header
		if bearerToken != "" {
			connectReq.Header.Set("Authorization", bearerToken) // "Bearer "+bearerToken)

		}

		//log.Printf("Headers: %v", connectReq.Header)

		if err := connectReq.Write(proxyConn); err != nil {
			proxyConn.Close()
			return nil, fmt.Errorf("failed to write HTTP CONNECT request: %w", err)
		}

		// Read the HTTP CONNECT response from the proxy server
		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connectReq)
		if err != nil {
			proxyConn.Close()
			return nil, fmt.Errorf("failed to read HTTP CONNECT response: %w", err)
		}

		//log.Printf("Response headers: %v", resp.Header)
		//bodyBytes, _ := ioutil.ReadAll(resp.Body)
		//log.Printf("Response body: %s", string(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			proxyConn.Close()

			return nil, fmt.Errorf("non-200 status code from proxy server: %d", resp.StatusCode)
		}

		return proxyConn, nil
	}
}
