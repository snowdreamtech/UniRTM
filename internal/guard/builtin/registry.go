// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin

import (
	"github.com/snowdreamtech/unirtm/internal/guard"
)

// GetAllBuiltinRules returns instances of all Go-native built-in guard rules
func GetAllBuiltinRules() []guard.Rule {
	return []guard.Rule{
		NewTrailingWhitespaceRule(),
		NewEndOfFileRule(),
		NewUtf8BOMRule(),
		NewMergeConflictRule(),
		NewPrivateKeyRule(),
		NewEnvSafetyRule(),
		NewLargeFilesRule(),
		NewCaseConflictRule(),
		NewLockfileCoexistenceRule(),
		NewLockfileConflictsRule(),
	}
}
