package cloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"math/rand"
	"os"
	"strings"
	"time"
)

type S3config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type S3archive struct {
	client     *s3.S3
	bucket     string
	randSource *rand.Rand
}

func NewS3archive(config S3config) (S3archive, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return S3archive{}, err
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return S3archive{
		client:     s3.New(newSession),
		bucket:     config.Bucket,
		randSource: r,
	}, nil
}

func (s *S3archive) SendFile(path string) error {
	logfile, err := os.Open(path)
	if err != nil {
		return err
	}

	split := strings.Split(path, "/")
	name := split[len(split)-2]
	now := time.Now().Format("2006_01_02_15_04_05")

	key := fmt.Sprintf("%s/%s_%d", name, now, s.randSource.Int())
	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   logfile,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *S3archive) DeleteFolder(path string) error {
	split := strings.Split(path, "/")

	key := ""
	if len(split) >= 1 {
		key = split[len(split)-1]
	}

	list, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("test"),
		Prefix: aws.String(key),
	})

	if err != nil {
		return err
	}

	for _, obj := range list.Contents {
		_, err = s.client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    obj.Key,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
