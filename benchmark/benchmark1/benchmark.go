package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"log_rotation_tool/archive/cloud"
	"log_rotation_tool/rotator"
	"os"
	"time"
)

const (
	rotations = 100
)

var log = &logrus.Logger{
	Out:          os.Stderr,
	Formatter:    new(logrus.TextFormatter),
	Level:        logrus.InfoLevel,
	ReportCaller: true,
}

func main() {
	period := flag.Duration("period", 15*time.Second, "period")
	reduce := flag.Int("reduce", 2, "reduce")
	testDir := flag.String("logDir", "/var/log/pods", "log directories")
	rotateDir := flag.String("rotateDir", "/var/log/rotate", "directories for rotated files")

	flag.Parse()

	config := cloud.S3config{
		Endpoint:  os.Getenv("ENDPOINT"),
		Region:    os.Getenv("REGION"),
		Bucket:    os.Getenv("BUCKET"),
		AccessKey: os.Getenv("ACCESS_KEY"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	archive, err := cloud.NewS3archive(config)
	if err != nil {
		panic(err)
	}

	rot := rotator.Rotator{
		LogDirs:    []string{*testDir},
		RotateDir:  *rotateDir,
		Archive:    &archive,
		ReduceSize: 2000,
	}

	// clear all existing log files with not measured rotation
	if err := rot.ReduceFiles(); err != nil {
		panic(err)
	}
	err = rot.Rotate()
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(*rotateDir, 0777); err != nil {
		panic(err)
	}

	// start size reduction
	go func() {
		reducePeriodLength := (*period).Milliseconds() / int64(*reduce+1)
		reduceDuration, err := time.ParseDuration(fmt.Sprintf("%dms", reducePeriodLength))

		if err != nil {
			panic(err)
		}

		ticker := time.NewTicker(reduceDuration)
		finishedReducePeriods := 0

		for {
			select {
			case <-ticker.C:
				if finishedReducePeriods >= *reduce {
					finishedReducePeriods = 0
					break
				}

				if err := rot.ReduceFiles(); err != nil {
					log.WithError(err).Error("rotator.ReduceFiles failed")
				}

				finishedReducePeriods++
			}
		}
	}()

	ticker := time.NewTicker(*period)
	finishedRotations := 0
	totalScore := 0.

	// measure rotations
	for {
		if finishedRotations == rotations {
			break
		}

		select {
		case <-ticker.C:
			score, err := measuredRotation(rot)
			if err != nil {
				log.Printf("rotation failed with error %s\n", err.Error())
				continue
			}

			totalScore += score
			finishedRotations++
			log.Printf("finished rotation, finished: %d, last score: %f\n", finishedRotations, score)
		}
	}

	log.Printf("mean rotation time: %f\n", totalScore/rotations)

	time.Sleep(24 * time.Hour)
}

func measuredRotation(rotator rotator.Rotator) (float64, error) {
	start := time.Now()
	if err := rotator.Rotate(); err != nil {
		return 0, err
	}

	return time.Since(start).Seconds(), nil
}
