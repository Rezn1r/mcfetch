# mcfetch Installer for Windows

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\mcfetch"
)

$ErrorActionPreference = "Stop"

$repo = "Rezn1r/mcfetch"
$apiUrl = "https://api.github.com/repos/$repo/releases/latest"

Write-Host "mcfetch Installer" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan
Write-Host ""

try {
    # get latest release information
    Write-Host "Fetching latest release information..." -ForegroundColor Yellow
    $response = Invoke-RestMethod -Uri $apiUrl -Headers @{
        "User-Agent" = "mcfetch-installer"
    }
    
    if (-not $response -or -not $response.tag_name) {
        Write-Host "Error: Failed to fetch release information." -ForegroundColor Red
        exit 1
    }
    
    $version = $response.tag_name
    Write-Host "Latest version: $version" -ForegroundColor Green
    
    # find Windows executable - match mcfetch-windows-amd64.exe
    $asset = $response.assets | Where-Object { $_.name -eq "mcfetch-windows-amd64.exe" }
    
    if (-not $asset) {
        # fallback: try common patterns
        $asset = $response.assets | Where-Object { $_.name -like "*windows*.exe" -or $_.name -like "*win*.exe" }
    }
    
    if (-not $asset) {
        # last resort: any .exe file
        $asset = $response.assets | Where-Object { $_.name -like "*.exe" }
    }
    
    if (-not $asset) {
        Write-Host "Error: No Windows executable found in latest release." -ForegroundColor Red
        Write-Host "Available assets:" -ForegroundColor Yellow
        $response.assets | ForEach-Object { Write-Host "  - $($_.name)" }
        exit 1
    }
    
    $downloadUrl = $asset.browser_download_url
    $fileName = $asset.name
    
    Write-Host "Found asset: $fileName" -ForegroundColor Green
    Write-Host ""
    
    # create installation directory
    if (-not (Test-Path $InstallPath)) {
        Write-Host "Creating installation directory: $InstallPath" -ForegroundColor Yellow
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    $exePath = Join-Path $InstallPath "mcfetch.exe"
    
    # dl the executable
    Write-Host "Downloading $fileName..." -ForegroundColor Yellow
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing
    $ProgressPreference = 'Continue'
    
    Write-Host "Downloaded successfully!" -ForegroundColor Green
    Write-Host ""
    
    # verif the file was downloaded
    if (-not (Test-Path $exePath)) {
        Write-Host "Error: Download failed." -ForegroundColor Red
        exit 1
    }
    
    $fileSize = (Get-Item $exePath).Length
    if ($fileSize -eq 0) {
        Write-Host "Error: Downloaded file is empty." -ForegroundColor Red
        Remove-Item $exePath -Force
        exit 1
    }
    
    Write-Host "Installed to: $exePath ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor Green
    
    # add to PATH automatically
    Write-Host ""
    Write-Host "Adding to user PATH..." -ForegroundColor Yellow
    
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($userPath -notlike "*$InstallPath*") {
        $newPath = "$userPath;$InstallPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added $InstallPath to PATH" -ForegroundColor Green
        Write-Host "Please restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
    } else {
        Write-Host "$InstallPath is already in PATH" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "Installation complete! ✓" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    Write-Host "  mcfetch java donutsmp.net" -ForegroundColor White
    Write-Host "  mcfetch bedrock demo.mcstatus.io --verbose" -ForegroundColor White
    
} catch {
    Write-Host "" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Installation failed." -ForegroundColor Red
    exit 1
}
