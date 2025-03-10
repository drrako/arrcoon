#!/bin/bash

APP_NAME="arrcoon"
PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64" "darwin/amd64" "darwin/arm64")
BUILD_DIR="build"

# Get short Git commit SHA
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

cd ..
# mkdir -p BUILD_DIR
mkdir -p $BUILD_DIR/artifacts

for PLATFORM in "${PLATFORMS[@]}"; do
    OS=${PLATFORM%/*}
    ARCH=${PLATFORM#*/}
    OUTPUT="$BUILD_DIR/$APP_NAME-${OS}-${ARCH}"

    # Add .exe extension for Windows builds
    if [ "$OS" == "windows" ]; then
        OUTPUT+=".exe"
    fi

    echo "🚀 Building for $OS/$ARCH..."
    env GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w" -o "$OUTPUT" .

    if [ $? -ne 0 ]; then
        echo "❌ Failed to build for $OS/$ARCH"
        exit 1
    else
        echo "✅ Built: $OUTPUT"
    fi

    if [ "$OS" != "darwin" ] || [ "$ARCH" != "amd64" ]; then
        echo "Compressing binary with UPX..."
        upx --best --lzma "$OUTPUT"
        if [ $? -ne 0 ]; then
            echo "❌ UPX compression failed for $OUTPUT"
        else
            echo "Compressed: $OUTPUT"
        fi
    else
        echo "⏩ Skipping UPX for $OS/$ARCH"
    fi

    # Create a ZIP archive with the Git SHA tag
    ZIP_NAME="$BUILD_DIR/artifacts/$APP_NAME-${OS}-${ARCH}-${GIT_SHA}.zip"
    zip -j "$ZIP_NAME" "$OUTPUT"

    if [ $? -ne 0 ]; then
        echo "❌ Failed to create ZIP: $ZIP_NAME"
        exit 1
    else
        echo "✅ Created ZIP archive: $ZIP_NAME"
    fi
done