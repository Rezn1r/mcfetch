package main

// mcfetch - a cli tool to fetch and display Minecraft server status
// usage: mcfetch <edition> <host> [port]

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	box "github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/mcstatus-io/go-mcstatus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	flagArgs, positional := splitArgs(os.Args[1:])

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Print extra details")
	noColor := fs.Bool("no-color", false, "Disable colorized output")
	dryRun := fs.Bool("dry-run", false, "Show what would be fetched without calling the API")

	fs.Usage = func() {
		printHelp(fs)
	}

	if err := fs.Parse(flagArgs); err != nil {
		os.Exit(2)
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
	colorName := "Cyan"
	if color.NoColor {
		colorName = "White"
	}
	b := box.New(box.Config{Px: 1, Py: 0, Type: "Bold", Color: colorName, TitlePos: "Top"})
	b.Print(title, strings.Join(lines, "\n"))
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
