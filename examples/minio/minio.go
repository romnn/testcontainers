package main

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	tc "github.com/romnn/testcontainers"
	tcminio "github.com/romnn/testcontainers/minio"
)

func main() {
	container, err := tcminio.Start(context.Background(), tcminio.Options{
		// you could use latest here
		ImageTag:     "RELEASE.2022-10-20T00-55-09Z",
		RootUser:     "3846587325",
		RootPassword: "te782tcb7tr3va7brkwev7awst",
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// start logger
	logger, err := tc.StartLogger(context.Background(), container.Container)
	if err != nil {
		log.Printf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	// connect to minio
	uri := container.ConnectionURI()
	client, err := minio.New(uri, &minio.Options{
		Creds:  credentials.NewStaticV4(container.RootUser, container.RootPassword, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("failed to create client (%s): %v", uri, err)
	}

	// create a bucket
	bucket := "test-bucket"
	opts := minio.MakeBucketOptions{}
	err = client.MakeBucket(context.Background(), bucket, opts)
	if err != nil {
		log.Fatalf("failed to create bucket %q: %v", bucket, err)
	}
	log.Printf("created bucket %q", bucket)
}
