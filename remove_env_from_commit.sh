#!/bin/bash

# Script to remove .env file from a specific git commit
# Usage: ./remove_env_from_commit.sh <commit-hash>

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 <commit-hash>"
    echo ""
    echo "This script removes the .env file from a specific git commit and rewrites history."
    echo ""
    echo "Arguments:"
    echo "  commit-hash    The hash of the commit to modify"
    echo ""
    echo "Examples:"
    echo "  $0 d7e0d3dbd1b1e15700b1c2b099da4075fd82848b"
    echo "  $0 HEAD~5"
    echo ""
    echo "WARNING: This operation rewrites git history. Make sure you understand the implications."
}

# Check if commit hash is provided
if [ $# -eq 0 ]; then
    print_error "No commit hash provided."
    show_usage
    exit 1
fi

COMMIT_HASH=$1

# Validate that the commit exists
if ! git cat-file -t "$COMMIT_HASH" >/dev/null 2>&1; then
    print_error "Commit '$COMMIT_HASH' does not exist in this repository."
    exit 1
fi

print_info "Starting .env removal process for commit: $COMMIT_HASH"

# Check if .env exists in the specified commit
if ! git show "$COMMIT_HASH:.env" >/dev/null 2>&1; then
    print_warning ".env file does not exist in commit $COMMIT_HASH"
    print_info "No action needed."
    exit 0
fi

print_info ".env file found in commit $COMMIT_HASH"

# Show current status
print_info "Current git status:"
git status --porcelain
if [ -n "$(git status --porcelain)" ]; then
    print_error "Working directory is not clean. Please commit or stash your changes first."
    exit 1
fi

# Confirm action
echo ""
print_warning "This will rewrite git history and change commit hashes."
print_warning "Make sure you have:"
print_warning "  - Backed up any important data"
print_warning "  - Informed other collaborators"
print_warning "  - No uncommitted changes"
echo ""
read -p "Do you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Operation cancelled."
    exit 0
fi

# Perform the filter-branch operation
print_info "Removing .env from commit history..."
if git filter-branch --index-filter 'git rm --cached --ignore-unmatch .env' --prune-empty "$COMMIT_HASH"^..HEAD; then
    print_success ".env file successfully removed from commit history!"
else
    print_error "Failed to remove .env from commit history."
    exit 1
fi

# Check if .env exists in working directory and remove it
if [ -f ".env" ]; then
    print_warning ".env file exists in working directory."
    read -p "Do you want to remove it from working directory too? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm .env
        print_success ".env removed from working directory."
    fi
fi

# Offer to add .env to .gitignore
if [ ! -f ".gitignore" ]; then
    print_warning ".gitignore file does not exist."
    read -p "Do you want to create .gitignore and add .env to it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ".env" > .gitignore
        print_success ".gitignore created with .env entry."
    fi
else
    if ! grep -q "^\.env$" .gitignore; then
        print_warning ".env is not in .gitignore."
        read -p "Do you want to add .env to .gitignore? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ".env" >> .gitignore
            print_success ".env added to .gitignore."
        fi
    else
        print_info ".env is already in .gitignore."
    fi
fi

# Show final status
echo ""
print_info "Final git status:"
git status

echo ""
print_warning "IMPORTANT NEXT STEPS:"
print_warning "1. Review the rewritten history: git log --oneline"
print_warning "2. Force push to remote: git push --force-with-lease origin <branch-name>"
print_warning "3. Inform collaborators to reset their local branches"
print_warning "4. Create a new .env file with your configuration if needed"

print_success "Script completed successfully!"

