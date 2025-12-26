#!/bin/bash
# Validate all CALM architecture files in the architectures directory
for file in architectures/*.json; do
  echo "Validating $file..."
  calm validate -a "$file"
  if [ $? -ne 0 ]; then
    echo "Validation failed for $file"
    exit 1
  fi
done
echo "All architectures are valid."
