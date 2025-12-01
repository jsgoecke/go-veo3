#!/bin/bash
# Create minimal test fixtures for validation tests

# Create minimal valid JPEG (1x1 pixel)
echo "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wAALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAAA//EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AH//Z" | base64 -d > test.jpg

# Create minimal valid PNG (1x1 pixel)
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==" | base64 -d > test.png

# Create minimal valid WebP (1x1 pixel)
echo "UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEAAwA0JaQAA3AA/vuUAAA=" | base64 -d > test.webp

# Create an unsupported GIF file (1x1 pixel)
echo "R0lGODlhAQABAAAAACw=" | base64 -d > test.gif

# Create a 1KB file
dd if=/dev/zero of=1kb.jpg bs=1024 count=1 2>/dev/null

# Create a 1MB file
dd if=/dev/zero of=1mb.jpg bs=1048576 count=1 2>/dev/null

# Create a 10MB file
dd if=/dev/zero of=10mb.jpg bs=1048576 count=10 2>/dev/null

# Create exactly 20MB file
dd if=/dev/zero of=20mb.jpg bs=1048576 count=20 2>/dev/null

# Create 20MB + 1 byte file
dd if=/dev/zero of=20mb_plus_1.jpg bs=1 count=20971521 2>/dev/null

# Create a 21MB file
dd if=/dev/zero of=21mb.jpg bs=1048576 count=21 2>/dev/null

# Create a 100MB file (for testing large files)
dd if=/dev/zero of=100mb.jpg bs=1048576 count=100 2>/dev/null

# Create a valid MP4 file (minimal)
echo "AAAAIGZ0eXBpc29tAAACAGlzb21pc28yYXZjMW1wNDEAAAAIZnJlZQAAAMptZGF0AAAC" | base64 -d > test.mp4

# Create a second test image for interpolation
cp test.jpg test2.jpg

echo "Test fixtures created successfully!"