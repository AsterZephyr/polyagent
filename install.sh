#!/bin/bash
# PolyAgent Simple Installation Script

set -euo pipefail

echo "Installing PolyAgent..."

# Check Python version
if ! python3 --version | grep -q "3.1[1-9]"; then
    echo "Error: Python 3.11+ required"
    exit 1
fi

# Install dependency
pip3 install httpx

# Create .env if not exists
if [[ ! -f .env ]]; then
    cp .env.example .env
    echo "Created .env file from template"
    echo "Please edit .env and add your API keys"
fi

# Make main.py executable
chmod +x main.py

echo "Installation completed!"
echo ""
echo "Next steps:"
echo "1. Edit .env and add your API keys"
echo "2. Run: python3 main.py"