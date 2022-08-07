package rotator

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.InfoLevel,
}

type Rotator struct {
	LogDirs []string
	Archive Archive
}

func (r *Rotator) Rotate() error {
	files, err := r.listFiles()
	if err != nil {
		return fmt.Errorf("rotator.listFiles failed: %w", err)
	}

	for _, file := range files {
		if err := r.Archive.SendFile(file); err != nil {
			log.WithError(err).WithField("file", file).Error("failed to send file to Archive")
			continue
		}

		err := os.Remove(file)
		if err != nil {
			log.WithError(err).WithField("file", file).Error("os.Remove file failed")
			continue
		}

		log.WithField("file", file).Info("log file rotated")
	}

	return nil
}

func (r *Rotator) listFiles() ([]string, error) {
	var files []string

	for _, dir := range r.LogDirs {
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				log.WithField("path", dir).Warn("directory with log files does not exist")
				continue
			}

			return nil, fmt.Errorf("os.Stat(%s) failed: %w", dir, err)
		}

		if !info.IsDir() {
			log.WithField("path", dir).Warn("path is not a directory")
			continue
		}

		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.WithError(err).Infof("skipping invalid path: %s", path)

				return fs.SkipDir
			}

			if d.IsDir() {
				return nil
			}

			absolutePath, err := filepath.Abs(path)
			if err != nil {
				log.WithError(err).Errorf("failed to read absolute path from %s", path)

				return nil
			}

			files = append(files, absolutePath)

			return nil
		})

		if err != nil {
			log.WithField("path", dir).Warn("failed to list files in directory")
			continue
		}
	}

	return files, nil
}

type Archive interface {
	SendFile(filepath string) error
}
