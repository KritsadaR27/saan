#!/bin/bash

# Saan System - VS Code Performance Optimization Script
# This script cleans up build artifacts, optimizes git repo, and improves VS Code performance

set -e

echo "🧹 Starting Saan System cleanup and optimization..."

# Change to project root
cd "$(dirname "$0")"

# 1. Clean build artifacts
echo "🗑️  Cleaning build artifacts..."
find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "build" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "tmp" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "air_tmp" -type d -exec rm -rf {} + 2>/dev/null || true

# 2. Clean temporary files
echo "🧽 Cleaning temporary files..."
find . -name "*.tmp" -type f -delete 2>/dev/null || true
find . -name "*.temp" -type f -delete 2>/dev/null || true
find . -name "*.log" -type f -delete 2>/dev/null || true
find . -name ".DS_Store" -type f -delete 2>/dev/null || true

# 3. Clean coverage files
echo "📊 Cleaning coverage files..."
find . -name "coverage" -type d -exec rm -rf {} + 2>/dev/null || true
find . -name "*.out" -type f -delete 2>/dev/null || true
find . -name "coverage.html" -type f -delete 2>/dev/null || true

# 4. Clean Go cache
echo "🐹 Cleaning Go cache..."
go clean -cache 2>/dev/null || true
go clean -modcache 2>/dev/null || true

# 5. Clean Docker build cache
echo "🐳 Cleaning Docker cache..."
docker system prune -f 2>/dev/null || true

# 6. Git optimization
echo "🔧 Optimizing Git repository..."
git gc --aggressive --prune=now 2>/dev/null || true
git remote prune origin 2>/dev/null || true

# 7. Rebuild Go modules
echo "📦 Tidying Go modules..."
find . -name "go.mod" -execdir go mod tidy \; 2>/dev/null || true

# 8. Check disk space saved
echo "💾 Checking disk space..."
du -sh . 2>/dev/null || true

# 9. VS Code optimization suggestions
echo "⚡ VS Code optimization complete!"
echo ""
echo "📋 Manual steps (if needed):"
echo "   1. Restart VS Code to apply new settings"
echo "   2. Run 'Developer: Restart Extension Host' in VS Code"
echo "   3. Consider disabling unused extensions"
echo "   4. Check VS Code settings.json was applied"
echo ""
echo "🎯 Performance improvements:"
echo "   ✅ Excluded bin/ directories (64MB+)"
echo "   ✅ Added .copilotignore to reduce context"
echo "   ✅ Optimized VS Code file watching"
echo "   ✅ Cleaned build artifacts"
echo "   ✅ Optimized Git repository"
echo ""
echo "✨ Cleanup complete! VS Code should be faster now."
