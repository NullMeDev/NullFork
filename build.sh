#!/bin/bash

# Enhanced Gateway Scraper Build Script
# Creates both CLI and GUI versions from the same codebase

set -e

PROJECT_NAME="enhanced-gateway-scraper"
VERSION=$(cat VERSION 2>/dev/null || echo "v1.2.0")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR="./dist"
CMD_DIR="./cmd/gateway-scraper"
LDFLAGS="-X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitCommit=$GIT_COMMIT"

echo "🔨 Building Enhanced Gateway Scraper $VERSION"
echo "================================================"

# Create build directory
mkdir -p "$BUILD_DIR"

# Clean previous builds
rm -f "$BUILD_DIR"/*

# Build CLI version
echo "📦 Building CLI version..."
go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/gateway-scraper-cli" "$CMD_DIR"
echo "✅ CLI version built: $BUILD_DIR/gateway-scraper-cli"

# Build GUI version (same binary, different name for clarity)
echo "📦 Building GUI version..."
go build -ldflags "-X main.version=$VERSION" -o "$BUILD_DIR/gateway-scraper-gui" "$CMD_DIR"
echo "✅ GUI version built: $BUILD_DIR/gateway-scraper-gui"

# Create convenience symlinks/copies for backward compatibility
echo "🔗 Creating backward compatibility binaries..."
cp "$BUILD_DIR/gateway-scraper-cli" "$BUILD_DIR/gateway-scraper"
cp "$BUILD_DIR/gateway-scraper-cli" "$BUILD_DIR/scraper"
echo "✅ Backward compatibility binaries created"

# Create usage scripts
echo "📝 Creating usage scripts..."

# CLI launcher script
cat > "$BUILD_DIR/run-cli.sh" << 'EOF'
#!/bin/bash
# Enhanced Gateway Scraper - CLI Mode
echo "🚀 Starting Enhanced Gateway Scraper (CLI Mode)..."
./gateway-scraper-cli "$@"
EOF

# GUI launcher script  
cat > "$BUILD_DIR/run-gui.sh" << 'EOF'
#!/bin/bash
# Enhanced Gateway Scraper - GUI Mode
echo "🚀 Starting Enhanced Gateway Scraper (GUI Mode)..."
echo "📱 Web interface will be available at: http://localhost:8081"
./gateway-scraper-gui -gui "$@"
EOF

# Make scripts executable
chmod +x "$BUILD_DIR/run-cli.sh"
chmod +x "$BUILD_DIR/run-gui.sh"
chmod +x "$BUILD_DIR/gateway-scraper-cli"
chmod +x "$BUILD_DIR/gateway-scraper-gui"
chmod +x "$BUILD_DIR/gateway-scraper"
chmod +x "$BUILD_DIR/scraper"

echo "✅ Usage scripts created"

# Display build summary
echo ""
echo "🎉 Build completed successfully!"
echo "================================================"
echo ""
echo "📁 Built files in $BUILD_DIR:"
echo "  🖥️  gateway-scraper-cli     - CLI version"
echo "  🌐 gateway-scraper-gui     - GUI version (use with -gui flag)"
echo "  🔗 gateway-scraper         - Default (CLI, backward compatibility)"
echo "  🔗 scraper                 - Legacy name (backward compatibility)"
echo ""
echo "🚀 Quick start:"
echo "  CLI Mode:  $BUILD_DIR/run-cli.sh"
echo "  GUI Mode:  $BUILD_DIR/run-gui.sh"
echo ""
echo "🔧 Manual usage:"
echo "  CLI: $BUILD_DIR/gateway-scraper-cli"
echo "  GUI: $BUILD_DIR/gateway-scraper-gui -gui"
echo ""
echo "📍 GUI will be available at: http://localhost:8081"
