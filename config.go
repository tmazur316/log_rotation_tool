package main

import (
	"os"

	"log_rotation_tool/archive"
	"log_rotation_tool/archive/cloud"
	"log_rotation_tool/archive/fake"
)

func configureArchive(archiveType string) archive.Archive {
	switch archiveType {
	case "s3":
		return s3archive()
	default:
		return fake.NewArchive()
	}
}

func s3archive() archive.Archive {
	config := cloud.S3config{
		Endpoint:  os.Getenv("ENDPOINT"),
		Region:    os.Getenv("REGION"),
		Bucket:    os.Getenv("BUCKET"),
		AccessKey: os.Getenv("ACCESS_KEY"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	s3, err := cloud.NewS3archive(config)
	if err != nil {
		log.WithError(err).Error("received error during creation of s3 archive. Using fake archive instead")
		return fake.NewArchive()
	}

	return &s3
}
