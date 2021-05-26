#!/bin/bash

set -Eeuo pipefail

# Check if user has go installed on the system
if ! command -v go &> /dev/null
then
    echo "Golang installation not be found"
    exit
fi

LOCATION=$(pwd)
FILENAME="clean_registry"
DIRECTORY_PATH="$LOCATION/cmd"
FILEPATH="$DIRECTORY_PATH/$FILENAME.go"

# Check if installation already exists in the directory
if [ -f "$LOCATION/$FILENAME" ]
then
  echo 'Clean registry manager build already exist, do you want to overwrite it?'
  echo 'Enter y for Yes or n for No'
  read -r response
  if [ "$response" != "y" ]
  then
    exit
  fi
fi

# Check if the cli file exists in the directory
if [ ! -f "$FILEPATH" ]
then
  echo Clean registry manager script not found
  exit
fi

cd "$DIRECTORY_PATH"
go build -o "../cleanregistry"
