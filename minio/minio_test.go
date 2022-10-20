package minio

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	tc "github.com/romnn/testcontainers"
)

// TestMinio ...
func TestMinio(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, err := Start(ctx, Options{
		ImageTag:     "RELEASE.2022-10-20T00-55-09Z",
		RootUser:     "3846587325",
		RootPassword: "te782tcb7tr3va7brkwev7awst",
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// start logger
	logger, err := tc.StartLogger(ctx, container.Container)
	if err != nil {
		t.Errorf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	// connect to minio
	minioURI := container.ConnectionURI()
	minioClient, err := minio.New(minioURI, &minio.Options{
		Creds:  credentials.NewStaticV4(container.RootUser, container.RootPassword, ""),
		Secure: false,
	})
	if err != nil {
		t.Fatalf("failed to create client (%s): %v", minioURI, err)
	}

	// create a bucket
	bucket := "test-bucket"
	if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		// check if the bucket already exists
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if !(errBucketExists == nil && exists) {
			t.Errorf("failed to create bucket %q: %v", bucket, err)
		}
	}

	// insert some data
	objectID := "test-object"
	opts := minio.PutObjectOptions{ContentType: "text/plain"}
	reader := bytes.NewReader([]byte("hello world"))
	if _, err := minioClient.PutObject(ctx, bucket, objectID, reader, reader.Size(), opts); err != nil {
		t.Errorf("failed to put object: %v", err)
	}

	// get the data back
	object, err := minioClient.GetObject(ctx, bucket, objectID, minio.GetObjectOptions{})
	if err != nil {
		t.Errorf("failed to get object: %v", err)
	}
	info, err := object.Stat()
	if info.ContentType != "text/plain" {
		t.Errorf(`expected content type "text/plain" but got %s`, info.ContentType)
	}
	data, err := ioutil.ReadAll(object)
	if err != nil {
		t.Errorf("failed to read object: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf(`expected "hello world" but got %q`, string(data))
	}
}
