package rotator

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.DebugLevel,
}

var source = rand.NewSource(time.Now().UnixNano())

type Rotator struct {
	LogDirs    []string
	RotateDir  string
	Archive    Archive
	ReduceSize int64
}

func (r *Rotator) Rotate() error {
	err := filepath.WalkDir(r.RotateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.WithError(err).Infof("skipping invalid path: %s", path)

			return fs.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if err := r.Archive.SendFile(path); err != nil {
			log.WithError(err).WithField("file", path).Error("failed to send file to Archive")
		}

		err = os.Remove(path)
		if err != nil {
			log.WithError(err).WithField("file", path).Error("os.Remove file failed")
		}

		log.WithField("file", path).Info("log file rotated")

		return nil
	})

	if err != nil {
		return err
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

func (r *Rotator) ReduceFiles() error {
	files, err := r.listFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		split := strings.Split(file, "/")
		name := split[len(split)-2]
		rotatePath := fmt.Sprintf("%s/%s", r.RotateDir, name)

		info, err := os.Stat(file)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(rotatePath, os.ModePerm); err != nil {
			return err
		}

		if info.Size() >= r.ReduceSize {
			if err := os.Rename(file, fmt.Sprintf("%s/%d", rotatePath, source.Int63())); err != nil {
				return err
			}
		}
	}

	return nil
}

type Archive interface {
	SendFile(filepath string) error
	DeleteFolder(filepath string) error
}
