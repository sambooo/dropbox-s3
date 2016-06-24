package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	flag "github.com/spf13/pflag"
)

func main() {
	var (
		dir       string
		bucket    string
		bucketDir string
	)

	// Load .env
	err := godotenv.Load()
	handle(err)
	// Expand default path
	path, err := homedir.Expand("~/Dropbox/Screenshots")
	handle(err)

	// Setup flags
	flag.StringVar(&dir, "dir", path, "screenshot directory")
	flag.StringVar(&bucket, "bucket", "i.samby.co.uk", "s3 bucket")
	flag.StringVar(&bucketDir, "bucket-dir", "i", "directory to store screenshots in bucket")
	flag.Parse()

	// Find latest screenshot and store the extension
	filename, err := getLatestScreenshot(dir)
	handle(err)
	ext := filepath.Ext(filename)
	// Load its bytes and create a reader
	b, err := loadScreenshot(filename)
	handle(err)
	br := bytes.NewReader(b)
	// Get its trimmed sha256 hash to use as a filename
	trimmedHash := trimHash(getHash(b))
	// And make the output filename
	outname := makeFilename(bucketDir, trimmedHash, ext)

	// Create the S3 service
	service := s3.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	// Upload screenshot to S3
	err = uploadToS3(service, br, bucket, outname)
	handle(err)
}

func handle(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func getLatestScreenshot(dir string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files in dir: %q", dir)
	}
	return files[len(files)-1], nil
}

func loadScreenshot(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	return bytes, err
}

func getHash(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func trimHash(hash string) string {
	return hash[:12]
}

func makeFilename(bucketDir, hash, extension string) string {
	return filepath.Join(bucketDir, hash+extension)
}

func uploadToS3(service *s3.S3, file io.ReadSeeker, bucket, key string) error {
	fmt.Printf("uploading %q to %q\n", key, bucket)
	_, err := service.PutObject(&s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ContentType: aws.String(
			mime.TypeByExtension(filepath.Ext(key)),
		),
	})
	if err != nil {
		return err
	}
	fmt.Printf("uploaded http://%s\n", filepath.Join(bucket, key))
	return nil
}
