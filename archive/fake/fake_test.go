package fake_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"log_rotation_tool/archive/fake"
)

func TestArchive_SendFile(t *testing.T) {
	const logs = "example log entry"

	t.Run("should return error when file cannot be read", func(t *testing.T) {
		file := "/name/notExisting.txt"
		archive := fake.NewArchive()
		// when
		err := archive.SendFile(file)
		// then
		require.Error(t, err)
	})

	t.Run("should save new file in archive", func(t *testing.T) {
		file := fileWithLogs(t, logs)
		archive := fake.NewArchive()
		// when
		err := archive.SendFile(file)
		// then
		require.NoError(t, err)
		// and
		assert.Equal(t, logs, archive.GetFile(file))
	})

	t.Run("should add logs to existing file in archive", func(t *testing.T) {
		file := fileWithLogs(t, logs)
		archive := fake.NewArchive()
		expected := logs + "\n" + logs

		err := archive.SendFile(file)
		require.NoError(t, err)
		// when
		err = archive.SendFile(file)
		// then
		require.NoError(t, err)
		// and
		assert.Equal(t, expected, archive.GetFile(file))
	})
}

func fileWithLogs(t *testing.T, logs string) string {
	dir := t.TempDir()
	file, err := ioutil.TempFile(dir, "logs.txt")
	require.NoError(t, err)

	_, err = file.Write([]byte(logs))
	require.NoError(t, err)

	return file.Name()
}
