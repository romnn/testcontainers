package main

import (
	"context"
	"bytes"
	"io/ioutil"

	tcminio "github.com/romnnn/testcontainers/minio"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

func run() string {
	log.SetLevel(log.InfoLevel)
	// Start minio container
	minioC, minioCfg, err := tcminio.StartMinioContainer(context.Background(), tcminio.ContainerOptions{
		AccessKeyID: "3846587325",
		SecretAccessKey: "te782tcb7tr3va7brkwev7awst",
	})
	if err != nil {
		log.Fatalf("Failed to start minio container: %v", err)
	}
	defer minioC.Terminate(context.Background())

	// Connect to minio
	minioURI := minioCfg.ConnectionURI()
	minioClient, err := minio.New(minioURI, &minio.Options{
		Creds:  credentials.NewStaticV4(minioCfg.AccessKeyID, minioCfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create minio client (%s): %v", minioURI, err)
	}

	ctx := context.Background()

	// Create a bucket
	bucket := "test-bucket"
	if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		// Check if the bucket already exists
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if !(errBucketExists == nil && exists) {
			log.Fatalf("failed to create bucket %q: %v", bucket, err)
		}
	}

	// Insert some data
	objectID := "test-object"
	opts := minio.PutObjectOptions{ContentType: "text/plain"}
	reader := bytes.NewReader([]byte("hello world"))
	if _, err := minioClient.PutObject(ctx, bucket, objectID, reader, reader.Size(), opts); err != nil {
		log.Fatalf("failed to save resource to minio: %v", err)
	}

	// Get the data back
	object, err := minioClient.GetObject(ctx, bucket, objectID, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("failed to get resource from minio: %v", err)
	}
	data, err := ioutil.ReadAll(object)
	if err != nil {
		log.Fatalf("failed to read resource from minio: %v", err)
	}
	return string(data)
}

func main() {
	log.Infof("Inserted and retrieved object with content %q from minio", run())
}
