// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanOutputFormats(t *testing.T) {
	var buf bytes.Buffer
	h := NewFormatter(FormatterOptions{Format: FormatHuman})
	h.SetWriter(&buf)

	h.Success("success msg")
	assert.Contains(t, buf.String(), "success msg")
	buf.Reset()

	h.Warning("warn msg")
	assert.Contains(t, buf.String(), "warn msg")
	buf.Reset()

	h.Error("error msg")
	assert.Contains(t, buf.String(), "error msg")
	buf.Reset()

	h.Data(map[string]interface{}{"key": "val"})
	assert.Contains(t, buf.String(), "val")
	buf.Reset()

	h.Table([]string{"A"}, [][]string{{"1"}})
	assert.Contains(t, buf.String(), "A")
	assert.Contains(t, buf.String(), "1")
	buf.Reset()
}

func TestJSONOutputFormats(t *testing.T) {
	var buf bytes.Buffer
	h := NewFormatter(FormatterOptions{Format: FormatJSON})
	h.SetWriter(&buf)

	h.Success("success msg")
	assert.Contains(t, buf.String(), "success msg")
	buf.Reset()

	h.Warning("warn msg")
	assert.Contains(t, buf.String(), "warn msg")
	buf.Reset()

	h.Error("error msg")
	assert.Contains(t, buf.String(), "error msg")
	buf.Reset()

	h.Data(map[string]interface{}{"key": "val"})
	assert.Contains(t, buf.String(), "val")
	buf.Reset()
}

func TestProgressIndicator_Coverage(t *testing.T) {
	var buf bytes.Buffer
	opts := ProgressOptions{
		Writer:         &buf,
		ShowPercentage: true,
		ShowBytes:      true,
		ShowSpeed:      true,
		Width:          10,
	}
	p := NewProgressIndicator(opts)

	// test render with data
	p.Start()
	p.SetMessage("hello")
	p.Update(50, 100)

	p.(*progressIndicator).render()
	output := buf.String()
	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "50.0%")
	assert.Contains(t, output, "50 B")

	// test spinner path
	buf.Reset()
	p.(*progressIndicator).total = 0
	p.(*progressIndicator).render()

	// test fail path
	p.Fail(errors.New("fake error"))
	assert.Contains(t, buf.String(), "fake error")

	// no-op coverage
	np := NewNoOpProgressIndicator()
	np.Start()
	np.Update(1, 2)
	np.SetMessage("hi")
	np.Finish()
	np.Fail(nil)
}
