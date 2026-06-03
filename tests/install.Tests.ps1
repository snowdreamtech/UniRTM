BeforeAll {
    $script:InstallScript = "$PSScriptRoot\..\install.ps1"
}

Describe "install.ps1" {
    It "Should run with -Help without errors" {
        $output = & $InstallScript -Help *>&1 | Out-String
        $output | Should -Match "Usage"
    }
}
