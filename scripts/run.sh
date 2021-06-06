#!/bin/bash

set -Eeuo pipefail

# Check if number of arguments used to run the script is 4
if [ $# != 4 ]
then
  echo 'Three arguments must be specified, Digitalocean token, registry name, max image count and maximum percentage allowed.'
  exit
fi

# Save all 3 arguments to variables
TOKEN=$1
REGISTRY=$2
COUNT=$3
PERCENTAGE_THRESHOLD=$4

# Check if the first argument is a string
case $TOKEN in
    ''|*[!a-zA-Z0-9]*) echo 'First argument must be a string' ;;
    *);;
esac

# Check if the second argument is a string
case $REGISTRY in
    ''|*[!a-zA-Z]*) echo 'Second argument must be a string' ;;
    *);;
esac

# Check if the third argument is a number
case $COUNT in
    ''|*[!0-9]*) echo 'Third argument must be a number' ;;
    *);;
esac

# Check if the fourth argument is a number
case $PERCENTAGE_THRESHOLD in
    ''|*[!0-9]*) echo 'Fourth argument must be a number' ;;
    *);;
esac

# Navigate to root directory
cd ..

# Get file location information
LOCATION=$(pwd)
FILEPATH="$LOCATION/cleanregistry"

# Check if the cli file exists in the directory
if [ ! -x "$FILEPATH" ]
then
  echo 'Clean registry manager executable file not found'
  echo 'run '\''build.sh'\'' to build the executable file'
  exit
fi

# Run script
"$FILEPATH" -token "$TOKEN" -registry "$REGISTRY" -count "$COUNT" -percentage "$PERCENTAGE_THRESHOLD"

# Remove created script file
rm -f "$FILEPATH"