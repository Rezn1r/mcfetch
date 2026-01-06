# mcfetch Installer for Windows

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\mcfetch",
    [switch]$AddToPath
)

$ErrorActionPreference = "Stop"

$repo = "Rezn1r/mcstatus"
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
    
    $version = $response.tag_name
    Write-Host "Latest version: $version" -ForegroundColor Green
    
    # find Windows executable in assets
    $asset = $response.assets | Where-Object { $_.name -like "*windows*.exe" -or $_.name -like "*win*.exe" -or $_.name -eq "mcfetch.exe" }
    
    if (-not $asset) {
        # try to find any .exe file
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
    Write-Host "Installed to: $exePath ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor Green
    
    if ($AddToPath) {
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
    }
    
    Write-Host ""
    Write-Host "Installation complete! ✓" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    
    if ($AddToPath) {
        Write-Host "  mcfetch java donutsmp.net" -ForegroundColor White
        Write-Host "  mcfetch bedrock demo.mcstatus.io --verbose" -ForegroundColor White
    } else {
        Write-Host "  $exePath java donutsmp.net" -ForegroundColor White
        Write-Host "  $exePath bedrock demo.mcstatus.io --verbose" -ForegroundColor White
        Write-Host ""
        Write-Host "Tip: Run with -AddToPath to add mcfetch to your PATH:" -ForegroundColor Yellow
        Write-Host "  .\installer.ps1 -AddToPath" -ForegroundColor White
    }
    
} catch {
    Write-Host ""
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
