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

	"go.pedge.io/env"

	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
)

type appEnv struct {
	Dir       string `env:"SCREENSHOT_DIR,default=~/Dropbox/Screenshots"`
	Bucket    string `env:"SCREENSHOT_BUCKET,default=i.samby.co.uk"`
	BucketDir string `env:"SCREENSHOT_BUCKET_DIR,default=i"`
}

func main() {
	// Load .env and ignore errors
	_ = godotenv.Load()
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	// Expand default path
	dir, err := homedir.Expand(appEnv.Dir)
	if err != nil {
		return err
	}
	// Find latest screenshot and store the extension
	filename, err := getLatestScreenshot(dir)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filename)
	// Load its bytes and create a reader
	b, err := loadScreenshot(filename)
	if err != nil {
		return err
	}
	br := bytes.NewReader(b)
	// Get its trimmed sha256 hash to use as a filename
	trimmedHash := trimHash(getHash(b))
	// And make the output filename
	outname := makeFilename(appEnv.BucketDir, trimmedHash, ext)

	// Create the S3 service
	service := s3.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	// Upload screenshot to S3
	err = uploadToS3(service, br, appEnv.Bucket, outname)
	if err != nil {
		return err
	}
	// Copy url
	err = copyUrl(appEnv.Bucket, outname)
	if err != nil {
		return err
	}
	return nil
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

func copyUrl(bucket, key string) error {
	err := clipboard.WriteAll(fmt.Sprintf("http://%s\n", filepath.Join(bucket, key)))
	if err == nil {
		fmt.Println("copied to clipboard")
	}
	return err
}
