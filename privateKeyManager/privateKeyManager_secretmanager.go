package privateKeyManager

import (
	"context"
	"fmt"
	"google.golang.org/genproto/protobuf/field_mask"
	"hash/crc32"
	"io"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// https://github.com/GoogleCloudPlatform/golang-samples/tree/c7914355b959d3265d86fe512971f1af5aec2e27/secretmanager

// [START secretmanager_add_secret_version]

// AddSecretVersion adds a new secret version to the given secret with the
// provided payload.
func AddSecretVersion(parent string, payloadAsString string) (secretVersionName string, err error) {
	// parent := "projects/my-project/secrets/my-secret"

	// Declare the payload to store.
	payload := []byte(payloadAsString)
	// Compute checksum, use Castagnoli polynomial. Providing a checksum
	// is optional.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(payload, crc32c))

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: parent,
		Payload: &secretmanagerpb.SecretPayload{
			Data:       payload,
			DataCrc32C: &checksum,
		},
	}

	// Call the API.
	response, err := client.AddSecretVersion(ctx, req)
	secretVersionName = response.Name
	if err != nil {
		return "", fmt.Errorf("failed to add secret version: %v", err)
	}

	return secretVersionName, nil
}

// [START secretmanager_get_secret_version]

// GetSecretVersion gets information about the given secret version. It does not
// include the payload data.
func GetSecretVersion(w io.Writer, name string) error {
	// name := "projects/my-project/secrets/my-secret/versions/5"
	// name := "projects/my-project/secrets/my-secret/versions/latest"

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.GetSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get secret version: %v", err)
	}

	fmt.Fprintf(w, "Found secret version %s with state %s\n",
		result.Name, result.State)
	return nil
}

// [END secretmanager_get_secret_version]

// [START secretmanager_get_secret]

// GetSecret gets information about the given secret. This only returns metadata
// about the secret container, not any secret material.
func GetSecret(w io.Writer, name string) error {
	// name := "projects/my-project/secrets/my-secret"

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.GetSecretRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.GetSecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get secret: %v", err)
	}

	replication := result.Replication.Replication
	fmt.Fprintf(w, "Found secret %s with replication policy %s\n", result.Name, replication)
	return nil
}

// [END secretmanager_get_secret]

// [START secretmanager_access_secret_version]

// AccessSecretVersion accesses the payload for the given secret version if one
// exists. The version can be a version number as a string (e.g. "5") or an
// alias (e.g. "latest").
func AccessSecretVersion(name string) (string, error) {
	// name := "projects/my-project/secrets/my-secret/versions/5"
	// name := "projects/my-project/secrets/my-secret/versions/latest"

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("Data corruption detected.")
	}

	// WARNING: Do not print the secret in a production executionEnvironment - this snippet
	// is showing how to access the secret material.
	return string(result.Payload.Data[:]), nil
}

// [END secretmanager_access_secret_version]

// [START secretmanager_update_secret]

// UpdateSecret updates the metadata about an existing secret.
func UpdateSecret(w io.Writer, name string) error {
	// name := "projects/my-project/secrets/my-secret"

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.UpdateSecretRequest{
		Secret: &secretmanagerpb.Secret{
			Name: name,
			Labels: map[string]string{
				"secretmanager": "rocks",
			},
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"labels"},
		},
	}

	// Call the API.
	result, err := client.UpdateSecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update secret: %v", err)
	}
	fmt.Fprintf(w, "Updated secret: %s\n", result.Name)
	return nil
}

// [END secretmanager_update_secret]

// [START secretmanager_destroy_secret_version]

// DestroySecretVersion destroys the given secret version, making the payload
// irrecoverable. Other secrets versions are unaffected.
func DestroySecretVersion(name string) error {
	// name := "projects/my-project/secrets/my-secret/versions/5"

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.DestroySecretVersionRequest{
		Name: name,
	}

	// Call the API.
	if _, err := client.DestroySecretVersion(ctx, req); err != nil {
		return fmt.Errorf("failed to destroy secret version: %v", err)
	}
	return nil
}

// [END secretmanager_destroy_secret_version]
