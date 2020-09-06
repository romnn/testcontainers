package minio

import (
	"context"
	"testing"
	"bytes"
	"io/ioutil"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TestMinioContainer ...
func TestMinioContainer(t *testing.T) {
	t.Parallel()
	// Start minio container
	minioC, minioCfg, err := StartMinioContainer(context.Background(), ContainerOptions{
		AccessKeyID: "3846587325",
		SecretAccessKey: "te782tcb7tr3va7brkwev7awst",
	})
	if err != nil {
		t.Fatalf("Failed to start minio container: %v", err)
	}
	defer minioC.Terminate(context.Background())

	// Connect to minio
	minioURI := minioCfg.ConnectionURI()
	minioClient, err := minio.New(minioURI, &minio.Options{
		Creds:  credentials.NewStaticV4(minioCfg.AccessKeyID, minioCfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		t.Fatalf("Failed to create minio client (%s): %v", minioURI, err)
	}

	ctx := context.Background()

	// Create a bucket
	bucket := "test-bucket"
	if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		// Check if the bucket already exists
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if !(errBucketExists == nil && exists) {
			t.Errorf("failed to create bucket %q: %v", bucket, err)
		}
	}

	// Insert some data
	objectID := "test-object"
	opts := minio.PutObjectOptions{ContentType: "text/plain"}
	reader := bytes.NewReader([]byte("hello world"))
	if _, err := minioClient.PutObject(ctx, bucket, objectID, reader, reader.Size(), opts); err != nil {
		t.Errorf("failed to save resource to minio: %v", err)
	}

	// Get the data back
	object, err := minioClient.GetObject(ctx, bucket, objectID, minio.GetObjectOptions{})
	if err != nil {
		t.Errorf("failed to get resource from minio: %v", err)
	}
	info, err := object.Stat()
	if info.ContentType != "text/plain" {
		t.Errorf(`expected test object from minio to have content type "text/plain" but got %s`, info.ContentType)
	}
	data, err := ioutil.ReadAll(object)
	if err != nil {
		t.Errorf("failed to read resource from minio: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected content of the stored resource to be %q but got %q", "hello world", string(data))
	}
}
