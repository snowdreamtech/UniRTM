<#
  Copyright (c) 2026 SnowdreamTech. All rights reserved.
  Licensed under the MIT License. See LICENSE file in the project root for full license information.

  Run Pester tests for install.ps1 using PesterConfiguration.
  This script is invoked by the 'test:powershell' unirtm task.
#>

$c = New-PesterConfiguration
$c.Run.Path = 'tests/install.Tests.ps1'
$c.Run.Exit = $true
Invoke-Pester -Configuration $c
