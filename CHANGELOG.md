# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-06

### Added
- Initial release of mcfetch
- Support for querying Minecraft Java Edition servers
- Support for querying Minecraft Bedrock Edition servers
- `--verbose` flag for detailed server information
- `--no-color` flag to disable colored output
- `--dry-run` flag for testing without making API calls
- Automatic port defaults (Java: 25565, Bedrock: 19132)
- Windows PowerShell installer script with automatic PATH configuration

### Features
- Simple CLI interface: `mcfetch <java|bedrock> <host> [port]`
- Beautiful formatted output with server information boxes
- Edition-specific status display (MOTD, players, version, etc.)
- Port validation (1-65535)
- Single binary with no dependencies

[1.0.0]: https://github.com/Rezn1r/mcstatus/releases/tag/v1.0.0
