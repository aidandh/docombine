#!/bin/bash

# Check if directory argument is provided
if [ $# -lt 2 ]; then
    echo "Usage: $0 <url> <directory>"
    exit 1
fi

# URL for the POST request
url="$1"

# Directory containing the documents
directory="$2"

# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Error: Directory '$directory' not found."
    exit 1
fi

# Initialize the curl command
curl_cmd="curl --request POST $url --output -"

# Loop through each file in the directory
for file_path in "$directory"/*; do
    # Check if the path points to a file (not directory)
    if [[ -f "$file_path" ]]; then
        # Append each document using --form parameter to the curl command
        curl_cmd="$curl_cmd -s --form documents=@\"$file_path\" > /dev/null"
    fi
done

# Execute the constructed curl command
eval "$curl_cmd"
