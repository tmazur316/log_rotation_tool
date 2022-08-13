package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"log_rotation_tool/rotator"
)

var log = &logrus.Logger{
	Out:          os.Stderr,
	Formatter:    new(logrus.TextFormatter),
	Level:        logrus.InfoLevel,
	ReportCaller: false,
}

func main() {
	dirs := flag.String("dirs", "/var/log/pods", "list of comma seperated paths to directories containing log files")
	period := flag.Duration("period", 10*time.Second, "log rotation period")
	archiveType := flag.String("archive", "fake", "archive type to use, available options: fake/s3 (default fake)")
	reduce := flag.Int("reduce", 1, "number of file size reduction periods between rotations")
	reduceSize := flag.Int("reduceSize", 5000, "minimum length of log file to reduce")
	rotateDir := flag.String("rotateDir", "/var/log/copy", "directories containing files for rotate")

	flag.Parse()

	flag.PrintDefaults()

	if !filepath.IsAbs(*rotateDir) {
		log.Error("rotation directory path is not absolute, stopping execution...")

		os.Exit(2)
	}

	if err := os.MkdirAll(*rotateDir, 0777); err != nil {
		log.Error("failed to create dir for rotated files, stopping execution...")

		os.Exit(2)
	}

	paths := strings.Split(*dirs, ",")

	var absolutePaths []string
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			log.WithField("path", path).Warn("skipping log directory: path not absolute")
			continue
		}

		absolutePaths = append(absolutePaths, path)
	}

	rot := rotator.Rotator{
		LogDirs:    absolutePaths,
		RotateDir:  *rotateDir,
		Archive:    configureArchive(*archiveType),
		ReduceSize: int64(*reduceSize),
	}
	stopRotationChan := make(chan bool, 1)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		startRotation(&rot, *period, stopRotationChan)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		startFileSizeReduction(&rot, *period, *reduce, stopRotationChan)
		wg.Done()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	stopRotationChan <- true
	wg.Wait()
}

func startRotation(rotator *rotator.Rotator, period time.Duration, stop chan bool) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	log.WithField("period", period).Info("starting log rotation")

	for {
		select {
		case <-ticker.C:
			if err := rotator.Rotate(); err != nil {
				log.WithError(err).Error("failed to rotate logs")
			}
		case <-stop:
			log.Info("stopping log rotation")
			return
		}
	}
}

func startFileSizeReduction(rotator *rotator.Rotator, rotationPeriod time.Duration, reducePeriods int, stop chan bool) {
	reducePeriodLength := rotationPeriod.Milliseconds() / int64(reducePeriods+1)
	reduceDuration, err := time.ParseDuration(fmt.Sprintf("%dms", reducePeriodLength))

	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(reduceDuration)
	finishedReducePeriods := 0

	for {
		select {
		case <-ticker.C:
			if finishedReducePeriods >= reducePeriods {
				finishedReducePeriods = 0
				break
			}

			if err := rotator.ReduceFiles(); err != nil {
				log.WithError(err).Error("rotator.ReduceFiles failed")
			}

			finishedReducePeriods++
		case <-stop:
			log.Info("stopping log files size reduction")
			return
		}
	}
}
