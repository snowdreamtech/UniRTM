// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"errors"
	"testing"
)

func TestNoOpProgressIndicator_More(t *testing.T) {
	pi := NewNoOpProgressIndicator()
	// None of these should panic or do anything
	pi.Start()
	pi.Update(10, 100)
	pi.SetMessage("msg")
	pi.Finish()
	pi.Fail(errors.New("err"))
}

func TestSuggest(t *testing.T) {
	var buf bytes.Buffer
	Suggest(&buf, "instal", []string{"install", "uninstall", "init", "list"})
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("install")) {
		t.Errorf("expected suggestion 'install' in output, got: %s", output)
	}

	buf.Reset()
	Suggest(&buf, "completelywrong", []string{"install", "uninstall"})
	output = buf.String()
	if bytes.Contains([]byte(output), []byte("install")) {
		t.Errorf("did not expect 'install' for completelywrong string, got: %s", output)
	}
}
