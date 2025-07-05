#!/bin/bash

echo "🚀 Setting up Process Service..."

# Create project structure
echo "📁 Creating project structure..."
mkdir -p src

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Build the project
echo "🔨 Building project..."
npm run build

# Make CLI executable
echo "🔧 Setting up CLI..."
chmod +x dist/proc-service.js
chmod +x dist/proc-cli.js

echo "✅ Setup complete!"
echo ""
echo "🎯 Quick Start:"
echo "  1. Start the service: npm run proc:start"
echo "  2. In another terminal: npm run cli status"
echo "  3. Or visit: http://localhost:3737/status"
echo ""
echo "📚 For more info, see README.md"
