#!/bin/bash

set -Eeuo pipefail

# Set GO version
GO_VERSION="1.16.5"

# Set colors
RED='\033[31m'
BLUE='\033[34m'
GREEN='\033[32m'

cd ..

# Set file location information
LOCATION=$(pwd)
FILENAME="clean_registry"
FILEPATH="$LOCATION/cmd/$FILENAME.go"
GO_FILE_PATH="$(pwd)/cmd"

echo "
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|           CLEAN REGISTRY MANAGER          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
"

function validate_go_installation() {
  # Check if user has go installed on the system
  if ! command -v go &>/dev/null; then
    echo -e "${RED} Go installation not be found \033[0m"
    echo -e "${BLUE} Installing Go "
    sleep 2
    cd /tmp &&
      wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz &&
      sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

    export PATH=$PATH:/usr/local/go/bin
    go version

    echo -e "\033[0m"
    echo -e "${GREEN} Go has been installed! \033[0m"
  else
    echo -e "${GREEN} Go installation was found \033[0m"
  fi
}

function check_for_existing_build() {
  echo ""

  # Check if installation already exists in the directory
  if [ -f "$LOCATION/$FILENAME" ]; then
    echo -e "Clean registry manager build already exist, do you want to overwrite it?"
    echo "Enter y for Yes or n for No"
    read -r response
    sleep 2
    if [ "$response" != "y" ]; then
      echo ""
      echo "You have successfully cancelled the build process"
      exit
    fi
  fi
}

function validate_go_file() {
  echo ""

  # Check if the cli file exists in the directory
  if [ ! -f "$FILEPATH" ]; then
    echo -e "${RED} Clean registry manager script not found \033[0m"
    exit
  fi
}

function build_script() {
  # Check if go file exists
  validate_go_file

  # Navigate to CMD file directory containing go file
  cd "$GO_FILE_PATH"

  echo -e "${GREEN}
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |   Building Clean Registry Manager Script  |
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  \033[0m"
  sleep 2

  go build -o "../$FILENAME"

  echo -e "${GREEN}
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |                 Build Done                |
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  \033[0m"
}

# Check if user has go installed on the system
validate_go_installation

# Check if build exists
check_for_existing_build

build_script