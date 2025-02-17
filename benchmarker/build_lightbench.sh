#!/bin/bash

OUTPUT_DIR="./cmd/lightbench/build"
BUILD_TARGET="./cmd/lightbench/..."

platforms=("darwin/amd64" "darwin/arm64" "windows/amd64" "linux/amd64")

mkdir -p $OUTPUT_DIR
rm -rf $OUTPUT_DIR/*

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$OUTPUT_DIR/${GOOS}_${GOARCH}/lightbench

    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    echo "Building for $GOOS/$GOARCH..."
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name -ldflags="-s -w" -trimpath $BUILD_TARGET
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

echo "Build completed!"
