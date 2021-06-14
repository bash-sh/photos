#!/usr/bin/env bash

repo=${1:-'github.com/bash-sh/photos'}
platformlist=${2:-'linux/amd64'}
package=${3:-'photos'}

IFS='|' read -ra platforms <<< "$platformlist"

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o build/$output_name $repo
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
