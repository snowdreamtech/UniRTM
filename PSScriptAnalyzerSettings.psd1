@{
    # PSScriptAnalyzer Settings
    # This configuration file defines custom rules for PSScriptAnalyzer
    # Reference: https://github.com/PowerShell/PSScriptAnalyzer

    # Exclude rules that conflict with project standards
    ExcludeRules = @(
        # PSUseBOMForUnicodeEncodedFile conflicts with .editorconfig (charset = utf-8)
        # Project uses UTF-8 without BOM for all files to maintain consistency
        'PSUseBOMForUnicodeEncodedFile'
    )

    # Include default rules
    IncludeDefaultRules = $true
}
