#!/bin/bash

#To run this bash, it's required to have Go installed.

#Setting platforms
platforms=("windows/amd64/win-x64" "windows/386/win-x32" "linux/amd64/linux-x64" "linux/386/linux-x32")

#Setting project name and output name
name="logbackup"
main="main.go"

#Setting Folder Paths
binFld=$(pwd)
srcFld=$(cd ../src && pwd)

#Setting environment for each platform.
for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
	
	DEST="$binFld/${platform_split[2]}"
	PACKAGE="$srcFld/$main"
	
	output_name="$DEST/$name"

    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

	echo "Compiling $output_name"
	env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $PACKAGE

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
