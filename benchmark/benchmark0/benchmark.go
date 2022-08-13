package main

import (
	"fmt"
	"log"
	"log_rotation_tool/archive/cloud"
	"log_rotation_tool/rotator"
	"os"
	"strings"
)

const (
	testDir       = "/var/log/test"
	filename      = "test"
	triesPerSize  = 100
	filesPerTrial = 6
)

func main() {
	config := cloud.S3config{
		Endpoint:  "frfk2avjbygn.compat.objectstorage.eu-frankfurt-1.oraclecloud.com",
		Region:    "eu-frankfurt-1",
		Bucket:    "test",
		AccessKey: "b6089a0ea2dc94eb748f88c755ca0b42366de732",
		SecretKey: "jZh1l+BSKy1n/Cm659JdtkUAs/5DxS81YbmDteUqG3k=",
	}

	archive, err := cloud.NewS3archive(config)
	if err != nil {
		panic(err)
	}

	//fileSizes := []int{1_000, 10_000, 100_000, 1_000_000}
	//
	//if err := os.MkdirAll(testDir, 0777); err != nil {
	//	os.Exit(1)
	//}
	//
	rot := rotator.Rotator{
		RotateDir: testDir,
		Archive:   &archive,
	}
	//
	//var score []testScore
	//
	//for i := 5; i <= filesPerTrial; i++ {
	//	for _, size := range fileSizes {
	//		successfulTries := 0
	//		failedTries := 0
	//		totalRotationTime := float64(0)
	//
	//		for j := 0; j < triesPerSize; j++ {
	//			if err := createFiles(size, i); err != nil {
	//				failedTries++
	//				continue
	//			}
	//
	//			start := time.Now()
	//			err := rot.Rotate()
	//			rotationTime := time.Since(start).Seconds()
	//
	//			if err != nil {
	//				failedTries++
	//				continue
	//			}
	//
	//			successfulTries++
	//			totalRotationTime += rotationTime
	//
	//			log.Printf("benchmark state: file size: %d, files per trial: %d, succcessful: %d, failed: %d, total time: %f\n",
	//				size, i, successfulTries, failedTries, totalRotationTime)

	if err := rot.Archive.DeleteFolder(testDir); err != nil {
		log.Println("failed to delete test file")
	}
	//}

	//score = append(score, testScore{
	//	fileSize:          size,
	//	filesPerTrial:     i,
	//	count:             successfulTries,
	//	totalRotationTime: totalRotationTime,
	//})
	//}
	//}

	//for _, s := range score {
	//	log.Printf("final score: file size: %d, files per trial: %d, mean rotation time: %f\n",
	//		s.fileSize, s.filesPerTrial, s.totalRotationTime/float64(s.count))
	//}
	//
	//time.Sleep(24 * time.Hour)
}

func createFiles(size int, filesNumber int) error {
	for i := 1; i <= filesNumber; i++ {
		name := fmt.Sprintf("%s/%s_%d", testDir, filename, i)

		file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return err
		}

		if _, err := file.Write([]byte(strings.Repeat("a", size/i))); err != nil {
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

type testScore struct {
	fileSize          int
	filesPerTrial     int
	count             int
	totalRotationTime float64
}
