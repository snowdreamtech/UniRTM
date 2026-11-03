// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"testing"
)

func TestExecuteBinary(t *testing.T) {
	// test with non-existent binary
	ExecuteBinary("non-existent-tool", []string{"arg1"})
}
