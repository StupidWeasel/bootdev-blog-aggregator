#!/bin/bash
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <up/down>"
    exit 1
fi

if [[ "$1" != "up" && "$1" != "down" ]]; then
    echo "Error: Argument must be 'up' or 'down'."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: 'jq' is not installed. Please install jq and try again."
    exit 1
fi

if [ ! -f ~/.gatorconfig.json ]; then
    echo "~/.gatorconfig.json missing from homedir"
    exit 1
fi

DB_URL=$(jq -r '.db_url' ~/.gatorconfig.json)
if [[ -z "$DB_URL" || "$DB_URL" == "null" ]]; then
    echo "Error: 'db_url' not found or empty in ~/.gatorconfig.json."
    exit 1
fi

cd "$(dirname "$0")/schema" || { echo "Error: Failed to change directory to schema"; exit 1; }
goose postgres "$DB_URL" $1
