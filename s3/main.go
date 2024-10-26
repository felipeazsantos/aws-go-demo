package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const bucketName = "aws-demo-felipeazsantos-bucket"
const regionName = "sa-east-1"

func main() {
	var (
		s3Client *s3.Client
		err      error
		out      []byte
	)
	ctx := context.Background()

	if s3Client, err = initS3Client(ctx); err != nil {
		log.Fatalf("initS3Client error, %s", err)
	}

	if err = createS3Bucket(ctx, s3Client); err != nil {
		log.Fatalf("createS3Bucket error, %s", err)
	}

	if err = uploadToS3Bucket(ctx, s3Client); err != nil {
		log.Fatalf("uploadToS3Bucket error, %s", err)
	}

	fmt.Printf("Upload complete.\n")

	if out, err = downloadFromS3(ctx, s3Client); err != nil {
		log.Fatalf("downloadToS3Bucket error, %s", err)
	}

	fmt.Printf("Download complete: %s\n", string(out))
}

func initS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(regionName))
	if err != nil {
		return nil, fmt.Errorf("unable to load sdk config, %s", err)
	}

	return s3.NewFromConfig(cfg), nil
}

func createS3Bucket(ctx context.Context, s3Client *s3.Client) error {
	allBuckets, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("ListBuckets error, %s", err)
	}

	found := false
	for _, bucket := range allBuckets.Buckets {
		if *bucket.Name == bucketName {
			found = true
			break
		}
	}

	if !found {
		_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: regionName,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func uploadToS3Bucket(ctx context.Context, s3Client *s3.Client) error {
	uploader := manager.NewUploader(s3Client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("test.txt"),
		Body:   strings.NewReader("hello world!"),
	})
	if err != nil {
		return fmt.Errorf("upload error, %s", err)
	}

	return nil
}

func downloadFromS3(ctx context.Context, s3Client *s3.Client) ([]byte, error) {
	downloader := manager.NewDownloader(s3Client)
	buffer := manager.NewWriteAtBuffer([]byte{})
	numBytes, err := downloader.Download(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("test.txt"),
	})
	if err != nil {
		return nil, fmt.Errorf("download object error, %s", err)
	}

	if numBytesReceived := len(buffer.Bytes()); numBytes != int64(numBytesReceived) {
		return nil, fmt.Errorf("numbytes received doesn't match: %d vs %d", numBytes, numBytesReceived)
	}

	return buffer.Bytes(), nil
}
