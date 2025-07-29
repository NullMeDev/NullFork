# Enhanced Gateway Scraper - Unified CLI/GUI Version

A powerful gateway scraping tool that can run in both CLI (headless) and GUI (web interface) modes using the same codebase.

## 🚀 Quick Start

### Build the application
```bash
# Build both CLI and GUI versions
make build

# Or use the build script
./build.sh
```

### Run CLI Mode (Headless)
```bash
# Using the built binary
./dist/gateway-scraper-cli

# Or using make
make run-cli

# Or using the launcher script
./dist/run-cli.sh
```

### Run GUI Mode (Web Interface)
```bash
# Using the built binary
./dist/gateway-scraper-gui -gui

# Or using make
make run-gui

# Or using the launcher script
./dist/run-gui.sh
```

The GUI will be available at: http://localhost:8081

## 📁 Project Structure

```
enhanced-gateway-scraper/
├── cmd/
│   └── gateway-scraper/          # Unified main entry point
│       └── main.go               # Single binary for both CLI/GUI
├── internal/                     # Shared internal packages
│   ├── api/                      # API server logic
│   ├── config/                   # Configuration management
│   ├── database/                 # Database connectivity
│   ├── discord/                  # Discord integration
│   └── logger/                   # Logging utilities
├── dist/                         # Built binaries (created by build)
│   ├── gateway-scraper-cli       # CLI version
│   ├── gateway-scraper-gui       # GUI version (same binary)
│   ├── gateway-scraper           # Default (CLI, backward compatibility)
│   ├── scraper                   # Legacy name compatibility
│   ├── run-cli.sh               # CLI launcher script
│   └── run-gui.sh               # GUI launcher script
├── web/                          # Web templates and assets (for GUI)
├── build.sh                      # Build script
├── Makefile                      # Build automation
└── README-UNIFIED.md            # This file
```

## 🔧 Building

### Using Make (Recommended)
```bash
# Build both versions
make build

# Build only CLI
make build-cli

# Build only GUI
make build-gui

# Clean build artifacts
make clean

# Run tests
make test

# Show help
make help
```

### Using Build Script
```bash
# Make it executable
chmod +x build.sh

# Build both versions
./build.sh
```

### Manual Build
```bash
# CLI version
go build -o dist/gateway-scraper-cli ./cmd/gateway-scraper

# GUI version (same binary, different name for clarity)
go build -o dist/gateway-scraper-gui ./cmd/gateway-scraper
```

## 🎯 Usage

### CLI Mode Features
- Headless operation
- API server on configured port
- Discord integration
- Database connectivity
- Background processing
- Suitable for servers and automation

### GUI Mode Features  
- Web-based interface on port 8081
- Real-time dashboard
- Visual monitoring
- Interactive controls
- Same backend functionality as CLI

### Command Line Options
```bash
gateway-scraper [options]

Options:
  -gui      Run in GUI mode (web interface on port 8081)
  -version  Show version information
  -help     Show help message

Examples:
  gateway-scraper           # Run in CLI mode
  gateway-scraper -gui      # Run in GUI mode
```

## 🔄 Migration from Old Structure

The new unified structure maintains backward compatibility:

| Old Binary Name | New Equivalent | Description |
|----------------|----------------|-------------|
| `scraper` | `gateway-scraper-cli` | CLI mode (default) |
| `gateway-scraper` | `gateway-scraper-cli` | CLI mode |
| `gateway-scraper-gui` | `gateway-scraper-gui -gui` | GUI mode |

## 📦 Installation

### System Installation
```bash
# Install to /usr/local/bin (requires sudo)
make install

# Now you can run from anywhere
gateway-scraper-cli
gateway-scraper-gui -gui
```

### Local Installation
Just run the binaries from the `dist/` directory after building.

## 🛠️ Development

### Adding Features
Since both CLI and GUI modes share the same codebase:

1. Add your feature to the appropriate internal package
2. Update both `runCLIMode()` and `runGUIMode()` functions as needed
3. Rebuild with `make build`

### Testing
```bash
# Run all tests
make test

# Test CLI mode
make run-cli

# Test GUI mode
make run-gui
```

## 🔍 Key Benefits

1. **Single Codebase**: Both CLI and GUI use the same core logic
2. **Clear Separation**: Obvious distinction between CLI and GUI modes
3. **Backward Compatibility**: Old binary names still work
4. **Easy Building**: Simple make commands and build scripts
5. **Flexible Deployment**: Choose the right mode for your environment

## 📖 Configuration

Both CLI and GUI modes use the same configuration system. See the existing configuration documentation for details on:

- Environment variables
- Configuration files
- Database setup
- Discord integration

## 🤝 Contributing

1. Make changes to the shared internal packages
2. Test both CLI and GUI modes
3. Update documentation if needed
4. Ensure backward compatibility

## 📋 Available Make Commands

- `make build` - Build both CLI and GUI versions
- `make build-cli` - Build only CLI version
- `make build-gui` - Build only GUI version  
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make run-cli` - Build and run CLI version
- `make run-gui` - Build and run GUI version
- `make install` - Install to system
- `make help` - Show help

---

## Legacy Documentation

For information about the internal architecture, API endpoints, database schema, and other technical details, refer to the original README.md file.
