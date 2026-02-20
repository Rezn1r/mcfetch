package main

// mcfetch - a cli tool to fetch and display Minecraft server status
// usage: mcfetch <edition> <host> [port]

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	box "github.com/box-cli-maker/box-cli-maker/v3"
	"github.com/fatih/color"
	"github.com/mcstatus-io/go-mcstatus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	flagArgs, positional := splitArgs(os.Args[1:])
	versionNumber := "v1.2.1"
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Print extra details")
	version := fs.Bool("version", false, "Print version information and exit")
	noColor := fs.Bool("no-color", false, "Disable colorized output")
	dryRun := fs.Bool("dry-run", false, "Show what would be fetched without calling the API")
	update := fs.Bool("update", false, "Update mcfetch to the latest release")
	uninstall := fs.Bool("uninstall", false, "Uninstall mcfetch from this system")

	fs.Usage = func() {
		printHelp(fs)
	}

	if err := fs.Parse(flagArgs); err != nil {
		os.Exit(2)
	}

	// Handle update flag
	if *update {
		if err := handleUpdate(); err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		return
	}

	// Handle version flag
	if *version {
		fmt.Printf("mcfetch %s\n", versionNumber)
		return
	}

	// Handle uninstall flag
	if *uninstall {
		handleUninstall()
		return
	}

	if len(positional) < 2 {
		fs.Usage()
		os.Exit(2)
	}

	edition := strings.ToLower(strings.TrimSpace(positional[0]))
	host := positional[1]
	portStr := ""
	if len(positional) >= 3 {
		portStr = positional[2]
	}

	color.NoColor = *noColor

	port, err := parsePort(edition, portStr)
	if err != nil {
		printError(err.Error())
		os.Exit(2)
	}

	if *dryRun {
		lines := []string{
			formatField("Edition", edition),
			formatField("Host", fmt.Sprintf("%s:%d", host, port)),
			formatField("Mode", "dry run (no request)"),
		}
		printBox("mcfetch Dry Run", lines)
		return
	}

	switch edition {
	case "java":
		status, err := mcstatus.GetJavaStatus(host, port)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		printJava(host, port, status, *verbose)

	case "bedrock":
		status, err := mcstatus.GetBedrockStatus(host, port)
		if err != nil {
			printError(err.Error())
			os.Exit(1)
		}
		printBedrock(host, port, status, *verbose)

	default:
		printError("Unknown edition. Use 'java' or 'bedrock'")
		os.Exit(2)
	}
}

func printJava(host string, port uint16, status *mcstatus.JavaStatusResponse, verbose bool) {
	titleCaser := cases.Title(language.Und)
	online := false
	playersOnline, playersMax := 0, 0
	versionName := ""
	motd := ""

	if status != nil {
		online = status.Online
		if status.Players != nil {
			playersOnline = status.Players.Online
			playersMax = status.Players.Max
		}
		if status.Version != nil {
			versionName = valueOrFallback(status.Version.NameClean, status.Version.NameRaw)
		}
		motd = status.MOTD.Clean
	}

	lines := []string{
		formatField("Host", fmt.Sprintf("%s:%d", host, port)),
		formatField("IP", valueOrFallback(status.IPAddress, "unknown")),
		formatField("Status", statusText(online)),
		formatField("Players", fmt.Sprintf("%d / %d", playersOnline, playersMax)),
		formatField("Version", valueOrFallback(versionName, "unknown")),
		formatField("MOTD", valueOrFallback(oneLine(motd), "")),
	}

	if verbose && status != nil {
		lines = append(lines,
			formatField("EULA", titleCaser.String(boolText(status.EULABlocked))),
			formatField("Cache", fmt.Sprintf("%s → %s", status.RetrievedAt.Format(time.RFC3339), status.ExpiresAt.Format(time.RFC3339))),
			formatField("Mods", fmt.Sprintf("%d", len(status.Mods))),
			formatField("Plugins", fmt.Sprintf("%d", len(status.Plugins))),
		)
		if status.SRVRecord != nil {
			lines = append(lines, formatField("SRV", fmt.Sprintf("%s:%d", status.SRVRecord.Host, status.SRVRecord.Port)))
		}
	}

	printBox("Java Edition", lines)
}

func printBedrock(host string, port uint16, status *mcstatus.BedrockStatusResponse, verbose bool) {
	titleCaser := cases.Title(language.Und)
	online := false
	playersOnline, playersMax := 0, 0
	versionName := ""
	motd := ""
	edition := ""

	if status != nil {
		online = status.Online
		if status.Players != nil {
			playersOnline = status.Players.Online
			playersMax = status.Players.Max
		}
		if status.Version != nil {
			versionName = status.Version.Name
		}
		motd = status.MOTD.Clean
		edition = status.Edition
	}

	lines := []string{
		formatField("Host", fmt.Sprintf("%s:%d", host, port)),
		formatField("Status", statusText(online)),
		formatField("Players", fmt.Sprintf("%d / %d", playersOnline, playersMax)),
		formatField("Version", valueOrFallback(versionName, "unknown")),
		formatField("MOTD", valueOrFallback(oneLine(motd), "")),
		formatField("Edition", valueOrFallback(edition, "unknown")),
	}

	if verbose && status != nil {
		lines = append(lines,
			formatField("IP", valueOrFallback(status.IPAddress, "unknown")),
			formatField("EULA", titleCaser.String(boolText(status.EULABlocked))),
			formatField("Cache", fmt.Sprintf("%s → %s", status.RetrievedAt.Format(time.RFC3339), status.ExpiresAt.Format(time.RFC3339))),
			formatField("Mode", valueOrFallback(status.Gamemode, "")),
			formatField("ServerID", valueOrFallback(status.ServerID, "")),
		)
	}

	printBox("Bedrock Edition", lines)
}

func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func printError(msg string) {
	color.New(color.FgRed, color.Bold).Fprintf(os.Stderr, "Error: %s\n", msg)
}

func valueOrFallback(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func parsePort(edition string, portStr string) (uint16, error) {
	var defaultPort uint16

	switch edition {
	case "java":
		defaultPort = 25565
	case "bedrock":
		defaultPort = 19132
	default:
		return 0, fmt.Errorf("unknown edition: %s", edition)
	}

	if strings.TrimSpace(portStr) == "" {
		return defaultPort, nil
	}

	portNum, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}

	if portNum == 0 || portNum > 65535 {
		return 0, fmt.Errorf("port must be 1-65535")
	}

	return uint16(portNum), nil
}

func splitArgs(args []string) (flagArgs []string, positional []string) {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
			continue
		}
		positional = append(positional, a)
	}
	return
}

func printBox(title string, lines []string) {
	boxColor := box.Cyan
	if color.NoColor {
		boxColor = box.White
	}

	b := box.NewBox().
		Padding(1, 0).
		Style(box.Bold).
		Color(boxColor).
		TitlePosition(box.Top)

	fmt.Println(b.MustRender(title, strings.Join(lines, "\n")))
}

func formatField(label, value string) string {
	labelText := fmt.Sprintf("%-8s", label+":")
	if color.NoColor {
		return fmt.Sprintf("%s %s", labelText, value)
	}
	labelColor := color.New(color.FgYellow)
	return fmt.Sprintf("%s %s", labelColor.Sprint(labelText), value)
}

func statusText(online bool) string {
	if color.NoColor {
		if online {
			return "Online"
		}
		return "Offline"
	}
	if online {
		return color.New(color.FgGreen).Sprint("Online")
	}
	return color.New(color.FgRed).Sprint("Offline")
}

func printHelp(fs *flag.FlagSet) {
	tool := binName()
	usage := fmt.Sprintf("%s <java|bedrock> <host> [port] [flags]", tool)
	lines := []string{
		formatField("Usage", usage),
		formatField("Example", fmt.Sprintf("%s java donutsmp.net", tool)),
		formatField("Example", fmt.Sprintf("%s java donutsmp.net 25565 --no-color", tool)),
		formatField("Example", fmt.Sprintf("%s bedrock demo.mcstatus.io --verbose", tool)),
		formatField("Flags", ""),
	}

	fs.VisitAll(func(f *flag.Flag) {
		name := "--" + f.Name
		defaultVal := f.DefValue
		lines = append(lines, fmt.Sprintf("  %-14s %s (default: %s)", name, f.Usage, defaultVal))
	})

	printBox("mcfetch help", lines)
}

func binName() string {
	name := filepath.Base(os.Args[0])
	name = strings.TrimSuffix(name, ".exe")
	if name == "" {
		return "mcstatus"
	}
	return name
}

func handleUninstall() {
	exePath, err := os.Executable()
	if err != nil {
		printError(fmt.Sprintf("Failed to get executable path: %v", err))
		os.Exit(1)
	}

	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		printError(fmt.Sprintf("Failed to resolve executable path: %v", err))
		os.Exit(1)
	}

	fmt.Println(color.CyanString("mcfetch Uninstaller"))
	fmt.Println(color.CyanString("==================="))
	fmt.Println()

	fmt.Printf("%s Executable path: %s\n", color.YellowString("→"), exePath)
	fmt.Println()

	fmt.Print(color.YellowString("Are you sure you want to uninstall mcfetch? [y/N]: "))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		fmt.Println(color.YellowString("Uninstall cancelled."))
		return
	}

	fmt.Println()
	fmt.Println(color.YellowString("Removing mcfetch..."))

	// Remove the executable
	if err := os.Remove(exePath); err != nil {
		printError(fmt.Sprintf("Failed to remove executable: %v", err))
		os.Exit(1)
	}

	fmt.Println(color.GreenString("✓ Removed: %s", exePath))
	fmt.Println()
	fmt.Println(color.GreenString("Uninstallation complete!"))
	fmt.Println()
	fmt.Println(color.YellowString("Note: You may need to manually remove the directory from your PATH if it was added."))
}

// Update handling

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func handleUpdate() error {
	fmt.Println(color.CyanString("mcfetch Updater"))
	fmt.Println(color.CyanString("================"))

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	fmt.Printf("%s Current: %s\n", color.YellowString("→"), exePath)

	rel, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	assetURL, assetName := pickAssetForCurrentPlatform(rel)
	if assetURL == "" {
		var names []string
		for _, a := range rel.Assets {
			names = append(names, a.Name)
		}
		return fmt.Errorf("no compatible asset found for %s/%s. Available: %s", runtime.GOOS, runtime.GOARCH, strings.Join(names, ", "))
	}

	fmt.Printf("%s Latest: %s\n", color.YellowString("→"), rel.TagName)
	fmt.Printf("%s Downloading: %s\n", color.YellowString("→"), assetName)

	dir := filepath.Dir(exePath)
	tmpFile, err := os.CreateTemp(dir, "mcfetch-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer tmpFile.Close()

	if err := downloadToFile(assetURL, tmpFile); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err := tmpFile.Chmod(0o755); err != nil {
			return fmt.Errorf("failed to chmod new binary: %w", err)
		}
		// Replace in-place (safe on Unix; running inode stays mapped)
		if err := os.Rename(tmpPath, exePath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
		fmt.Println(color.GreenString("✓ Updated successfully to %s", rel.TagName))
		return nil
	}

	// Windows: running executable is locked. Use a PowerShell helper.
	newPath := exePath + ".new.exe"
	// Close the temp file and move into place as .new.exe in the same directory
	tmpFile.Close()
	if err := moveOrCopy(tmpPath, newPath); err != nil {
		return fmt.Errorf("failed to prepare new binary: %w", err)
	}

	psScript := `param([string]$Target, [string]$NewFile, [int]$Pid)
try { Wait-Process -Id $Pid -ErrorAction SilentlyContinue } catch {}
Start-Sleep -Milliseconds 200
Move-Item -Path $NewFile -Destination $Target -Force
Write-Output "Updated $Target"`

	psPath := filepath.Join(dir, "mcfetch-update.ps1")
	if err := os.WriteFile(psPath, []byte(psScript), 0o644); err != nil {
		return fmt.Errorf("failed to write helper script: %w", err)
	}

	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", psPath, "-Target", exePath, "-NewFile", newPath, "-Pid", fmt.Sprint(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start updater: %w", err)
	}

	fmt.Println(color.GreenString("✓ Update staged. The binary will be replaced after this process exits."))
	fmt.Println(color.YellowString("Note: If replacement fails due to permissions, run as Administrator."))
	return nil
}

func fetchLatestRelease() (*ghRelease, error) {
	const api = "https://api.github.com/repos/Rezn1r/mcfetch/releases/latest"
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mcfetch-updater")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api returned %s", resp.Status)
	}
	var rel ghRelease
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func pickAssetForCurrentPlatform(rel *ghRelease) (url string, name string) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	candidates := []string{
		fmt.Sprintf("mcfetch-%s-%s", osName, arch),
		fmt.Sprintf("%s-%s", osName, arch),
		osName,
	}
	for _, a := range rel.Assets {
		n := strings.ToLower(a.Name)
		for _, c := range candidates {
			if strings.Contains(n, c) {
				return a.URL, a.Name
			}
		}
		// Special-case windows .exe naming
		if osName == "windows" && (strings.HasSuffix(n, ".exe")) && strings.Contains(n, arch) {
			return a.URL, a.Name
		}
	}
	return "", ""
}

func downloadToFile(url string, out io.Writer) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mcfetch-updater")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

func moveOrCopy(src, dst string) error {
	// Try rename first (same volume)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback to copy
	rf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer rf.Close()
	wf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer wf.Close()
	if _, err := io.Copy(wf, rf); err != nil {
		return err
	}
	if err := wf.Close(); err != nil {
		return err
	}
	_ = os.Remove(src)
	return nil
}
