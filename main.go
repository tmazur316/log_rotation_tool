package main

import (
	"flag"
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
	Level:        logrus.DebugLevel,
	ReportCaller: true,
}

func main() {
	dirs := flag.String("dirs", "./copy3", "list of comma seperated paths to directories containing log files")
	period := flag.Duration("period", 10*time.Second, "log rotation period")
	archiveType := flag.String("archive", "fake", "archive type to use, available options: fake/s3 (default fake)")

	flag.Parse()

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
		LogDirs: absolutePaths,
		Archive: configureArchive(*archiveType),
	}
	stopRotationChan := make(chan bool, 1)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		startRotation(&rot, *period, stopRotationChan)
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
