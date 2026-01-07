# mcfetch

A fast and beautiful CLI tool to fetch and display Minecraft server status with styled output.

## Features

- Support for both **Java** and **Bedrock** editions
- Beautiful boxed output with color support
- Verbose mode for detailed server information
- Single binary, no dependencies required

## Installation

### Quick Install

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Rezn1r/mcfetch/main/installer.ps1 | iex
```

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/Rezn1r/mcfetch/main/installer.sh | bash
```

### From Source

```bash
git clone https://github.com/Rezn1r/mcfetch.git
cd mcfetch
go build -o mcfetch
```

## Usage

```bash
mcfetch <java|bedrock> <host> [port] [flags]
```

### Arguments

- `<edition>`: Server edition - `java` or `bedrock`
- `<host>`: Server hostname or IP address
- `[port]`: Server port (optional, uses edition default if omitted)

### Flags

- `--verbose`: Display extra details (EULA, cache, mods, plugins, SRV records)
- `--no-color`: Disable colorized output
- `--dry-run`: Show what would be fetched without making API calls

## Examples

### Basic Usage

Fetch Java server status (uses default port 25565):
```bash
mcfetch java donutsmp.net
```

Fetch Bedrock server status with custom port:
```bash
mcfetch bedrock demo.mcfetch.io 19132
```

### With Flags

Verbose output with extra details:
```bash
mcfetch java donutsmp.net --verbose
```

Clean output without colors:
```bash
mcfetch java donutsmp.net --no-color
```

Test without making API call:
```bash
mcfetch java donutsmp.net --dry-run
```

Combine multiple flags:
```bash
mcfetch bedrock play.example.com 19132 --verbose --no-color
```

## Output

### Java Edition Example

```
┏━ Java Edition ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Host:    donutsmp.net:25565                      ┃
┃ IP:      123.45.67.89                            ┃
┃ Status:  Online                                  ┃
┃ Players: 30369 / 50000                           ┃
┃ Version: Velocity 1.7.2-1.21.1                   ┃
┃ MOTD:    DonutSMP.net - Survival                 ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Verbose Mode

With `--verbose`, additional information is displayed:
- IP address
- EULA status
- Cache timestamps
- Number of mods
- Number of plugins
- SRV record (if available)

## Requirements

- Go 1.21 or higher (for building from source)

## Dependencies

- [box-cli-maker](https://github.com/Delta456/box-cli-maker) - Terminal box styling
- [fatih/color](https://github.com/fatih/color) - Color output
- [go-mcstatus](https://github.com/mcstatus-io/go-mcstatus) - Minecraft server status API

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## Acknowledgments

- Powered by [mcstatus.io API](https://mcstatus.io)