package fake

import (
	"os"
)

type Archive struct {
	logs map[string]string
}

func NewArchive() Archive {
	return Archive{logs: map[string]string{}}
}

func (a Archive) SendFile(path string) error {
	logs, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if prev, exists := a.logs[path]; exists {
		a.logs[path] = prev + "\n" + string(logs)

		return nil
	}

	a.logs[path] = string(logs)

	return nil
}

func (a Archive) GetFile(path string) string {
	return a.logs[path]
}
