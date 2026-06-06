<#
  Copyright (c) 2026 SnowdreamTech. All rights reserved.
  Licensed under the MIT License. See LICENSE file in the project root for full license information.
#>

BeforeAll {
    $script:InstallScript = "$PSScriptRoot\..\install.ps1"
}

Describe "install.ps1" {
    It "Should run with -Help without errors" {
        $output = & $InstallScript -Help *>&1 | Out-String
        $output | Should -Match "Usage"
    }

    It "Should bypass checksum with -SkipChecksum" {
        Mock Invoke-WebRequest {
            if ($Uri -match "api.github.com") { return '{ "tag_name": "v99.9.9" }' }
            if ($Uri -match "\.zip") {
                # Create dummy zip with unirtm.exe
                $zipPath = $OutFile
                Add-Type -AssemblyName System.IO.Compression.FileSystem
                $zipStream = [System.IO.File]::OpenWrite($zipPath)
                $zipArchive = New-Object System.IO.Compression.ZipArchive($zipStream, [System.IO.Compression.ZipArchiveMode]::Create)
                $entry = $zipArchive.CreateEntry("unirtm.exe")
                $writer = New-Object System.IO.StreamWriter($entry.Open())
                $writer.Write("dummy")
                $writer.Dispose()
                $zipArchive.Dispose()
                $zipStream.Dispose()
                return $true
            }
            return $true
        } -ModuleName $null

        $tmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ([System.Guid]::NewGuid().ToString()) -Force

        $output = & $InstallScript -InstallDir $tmpDir -SkipChecksum *>&1 | Out-String
        $output | Should -Match "Skipping checksum verification"

        Remove-Item -Recurse -Force $tmpDir
    }
}
