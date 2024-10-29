package main

import (
	"context"
	"fmt"
	"io"
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

type S3Client interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
}

type S3Uploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type S3Downloader interface {
	Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error)
}

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

	if err = uploadToS3Bucket(ctx, manager.NewUploader(s3Client)); err != nil {
		log.Fatalf("uploadToS3Bucket error, %s", err)
	}

	fmt.Printf("Upload complete.\n")

	if out, err = downloadFromS3(ctx, manager.NewDownloader(s3Client)); err != nil {
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

func createS3Bucket(ctx context.Context, s3Client S3Client) error {
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

func uploadToS3Bucket(ctx context.Context, uploader S3Uploader) error {
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

func downloadFromS3(ctx context.Context, downloader S3Downloader) ([]byte, error) {
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
