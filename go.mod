module github.com/jlambert68/FenixConnectorAdminShared

go 1.23.0

toolchain go1.24.2

//toolchain go1.21.5

require (
	cloud.google.com/go/pubsub v1.49.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/pat v1.0.2
	github.com/gorilla/sessions v1.2.2 // v1.3.0 gives session error when trying to get token from GCP
	github.com/jlambert68/FenixGrpcApi v0.0.0-20250610083226-66f4b211526b
	github.com/jlambert68/FenixScriptEngine v0.0.0-20241104143504-8f37e95bc346
	github.com/jlambert68/FenixSyncShared v0.0.0-20240911064419-da3d922610cb
	github.com/jlambert68/FenixTestInstructionsAdminShared v0.0.0-20241024135649-85f0f911fdda
	github.com/markbates/goth v1.80.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/net v0.40.0
	golang.org/x/oauth2 v0.30.0
	google.golang.org/api v0.233.0
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6

)

require (
	cloud.google.com/go/secretmanager v1.14.5
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb
)

require (
	cloud.google.com/go v0.120.0 // indirect
	cloud.google.com/go/auth v0.16.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.4.2 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-gota/gota v0.12.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.3 // indirect
	github.com/jackc/pgx/v4 v4.18.3 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250313205543-e70fdf4c4cb4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
)
