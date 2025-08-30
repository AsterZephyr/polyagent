#!/bin/bash
# PolyAgent Setup Script
# Following Unix conventions: Simple, idempotent, informative

set -e  # Exit on any error

echo "🤖 PolyAgent Setup"
echo "=================="

# Check Python version
echo "Checking Python version..."
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is required but not installed"
    exit 1
fi

PYTHON_VERSION=$(python3 -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
echo "✓ Python $PYTHON_VERSION found"

# Check required Python version
if [[ "$(printf '%s\n' "3.8" "$PYTHON_VERSION" | sort -V | head -n1)" != "3.8" ]]; then
    echo "❌ Python 3.8 or higher required, found $PYTHON_VERSION"
    exit 1
fi

# Install Python dependencies
echo -e "\nInstalling Python dependencies..."
if [ -f "requirements.txt" ]; then
    pip3 install -r requirements.txt
else
    # Minimal dependencies
    pip3 install httpx asyncio
    echo "✓ Basic dependencies installed"
fi

# Create config from example if not exists
echo -e "\nSetting up configuration..."
if [ ! -f "config/.env" ]; then
    cp config/env.example config/.env
    echo "✓ Created config/.env from example"
    echo "📝 Please edit config/.env and add your API keys"
else
    echo "✓ Configuration already exists"
fi

# Create docs directory structure
echo -e "\nSetting up document directories..."
mkdir -p docs/{medical,tech,general}
echo "✓ Document directories created"

# Test basic functionality
echo -e "\nTesting basic functionality..."
cd agent
if python3 test.py > /dev/null 2>&1; then
    echo "✓ Basic tests passed"
else
    echo "⚠️  Some basic tests failed (this may be normal if no API keys configured)"
fi
cd ..

# Check for API keys
echo -e "\nChecking API key configuration..."
source config/.env 2>/dev/null || true

api_keys_found=0
if [ ! -z "$OPENAI_API_KEY" ]; then
    echo "✓ OpenAI API key configured"
    ((api_keys_found++))
fi

if [ ! -z "$ANTHROPIC_API_KEY" ]; then
    echo "✓ Anthropic API key configured"
    ((api_keys_found++))
fi

if [ ! -z "$OPENROUTER_API_KEY" ]; then
    echo "✓ OpenRouter API key configured"
    ((api_keys_found++))
fi

if [ ! -z "$GLM_API_KEY" ]; then
    echo "✓ GLM API key configured"
    ((api_keys_found++))
fi

if [ $api_keys_found -eq 0 ]; then
    echo "❌ No API keys configured!"
    echo "📝 Please edit config/.env and add at least one API key"
    echo ""
    echo "Example:"
    echo "  export OPENAI_API_KEY='sk-your-key-here'"
    echo "  export ANTHROPIC_API_KEY='sk-ant-your-key-here'"
    exit 1
else
    echo "✅ Found $api_keys_found API key(s) configured"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Quick start:"
echo "  cd agent"
echo "  source ../config/.env"
echo "  python3 main.py"
echo ""
echo "Or test with a simple query:"
echo "  echo 'Hello, how are you?' | python3 main.py"