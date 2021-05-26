#!/bin/bash

set -Eeuo pipefail

# Check if number of arguments used to run the script is 2
if [ $# != 3 ]
then
  echo 'Three arguments must be specified, Digitalocean token, Registry name and count.'
  exit
fi

# Save all 3 arguments to variables
TOKEN=$1
REGISTRY=$2
COUNT=$3

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

LOCATION=$(pwd)
FILEPATH="$LOCATION/cleanregistry"

# Check if the cli file exists in the directory
if [ ! -x "$FILEPATH" ]
then
  echo 'Clean registry manager executable file not found'
  echo 'run '\''build.sh'\'' to build the executable file'
  exit
fi

"$FILEPATH" -token "$TOKEN" -registry "$REGISTRY" -count $COUNT

rm -f "$FILEPATH"