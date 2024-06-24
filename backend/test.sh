#!/bin/bash

# Check if directory argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <directory>"
    exit 1
fi

# Directory containing the documents
directory="$1"

# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Error: Directory '$directory' not found."
    exit 1
fi

# URL for the POST request
url="http://localhost:8080/combine"

# Initialize the curl command
curl_cmd="curl --request POST $url --output -"

# Loop through each file in the directory
for file_path in "$directory"/*; do
    # Check if the path points to a file (not directory)
    if [[ -f "$file_path" ]]; then
        # Append each document using --form parameter to the curl command
        curl_cmd="$curl_cmd --form documents=@\"$file_path\""
    fi
done

# Execute the constructed curl command
eval "$curl_cmd"
