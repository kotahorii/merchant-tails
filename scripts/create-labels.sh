#!/bin/bash

# Create GitHub labels for Dependabot
# Requires GitHub CLI (gh) to be installed and authenticated

echo "Creating GitHub labels for Dependabot..."

# Create 'dependencies' label
gh label create dependencies \
  --repo kotahorii/merchant-tails \
  --description "Pull requests that update a dependency file" \
  --color "0366d6" \
  --force

# Create 'go' label
gh label create go \
  --repo kotahorii/merchant-tails \
  --description "Pull requests related to Go code" \
  --color "00ADD8" \
  --force

# Create 'github-actions' label
gh label create github-actions \
  --repo kotahorii/merchant-tails \
  --description "Pull requests that update GitHub Actions code" \
  --color "000000" \
  --force

echo "Labels created successfully!"