$SecretsDir = "secrets"

$SecretsFiles = @(
    "db_app_ro_password.txt",
    "db_app_rw_password.txt",
    "db_super_password.txt"
)

if (-not (Test-Path $SecretsDir)) {
    New-Item -ItemType Directory -Path $SecretsDir | Out-Null
}

foreach ($file in $SecretsFiles) {
    $path = Join-Path $SecretsDir $file

    if (-not (Test-Path $path)) {

        $bytes = New-Object byte[] 24
        $rng = [System.Security.Cryptography.RNGCryptoServiceProvider]::Create()
        $rng.GetBytes($bytes)
        $rng.Dispose()

        $random = [Convert]::ToBase64String($bytes)

        Set-Content -Path $path -Value $random -NoNewline

        Write-Host "Created $path"
    }
}