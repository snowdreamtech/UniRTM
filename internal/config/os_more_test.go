package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOS_MockErrors(t *testing.T) {
	errMock := fmt.Errorf("mock os error")

	t.Run("ReadFileOrEmpty error", func(t *testing.T) {
		orig := OsReadFile
		defer func() { OsReadFile = orig }()
		OsReadFile = func(filename string) ([]byte, error) { return nil, errMock }

		data, err := ReadFileOrEmpty("some_path")
		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("LoadRawTOML read error", func(t *testing.T) {
		orig := OsReadFile
		defer func() { OsReadFile = orig }()
		OsReadFile = func(filename string) ([]byte, error) { return nil, errMock }

		node, err := LoadRawTOML("some_path")
		assert.Error(t, err)
		assert.Nil(t, node)
	})

	t.Run("SaveRawTOML write error", func(t *testing.T) {
		orig := OsWriteFile
		defer func() { OsWriteFile = orig }()
		OsWriteFile = func(filename string, data []byte, perm os.FileMode) error { return errMock }

		err := SaveRawTOML("some_path", nil)
		assert.Error(t, err)
	})

	t.Run("manager Load read error", func(t *testing.T) {
		orig := OsReadFile
		defer func() { OsReadFile = orig }()
		OsReadFile = func(filename string) ([]byte, error) { return nil, errMock }

		origStat := OsStat
		defer func() { OsStat = origStat }()
		OsStat = func(name string) (os.FileInfo, error) { return nil, nil }

		m := NewConfigManager()
		_, err := m.Load(context.Background(), "some_path")
		assert.Error(t, err)
	})

	t.Run("LoadGlobal read error", func(t *testing.T) {
		orig := OsReadFile
		defer func() { OsReadFile = orig }()
		OsReadFile = func(filename string) ([]byte, error) { return nil, errMock }

		_, err := LoadGlobal()
		assert.Error(t, err)
	})
}
