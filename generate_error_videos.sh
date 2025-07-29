#!/bin/bash

# Prompt for video directory if not provided
if [ -z "$VIDEO_DIR" ]; then
    read -p "Enter path for error videos (default: /etc/zurg/zurg.git/error_videos): " VIDEO_DIR
    VIDEO_DIR=${VIDEO_DIR:-"/etc/zurg/zurg.git/error_videos"}
fi

# Create video directory if it doesn't exist
mkdir -p "$VIDEO_DIR"

# Function to generate error video using ffmpeg's default font
generate_error_video() {
    local message="$1"
    local filename="$2"
    local output="$VIDEO_DIR/$filename"
    
    echo "Generating: $filename"
    
    ffmpeg -f lavfi -i color=c=black:s=1920x1080:d=10 \
        -vf "drawtext=text='$message':fontsize=72:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2" \
        -c:v libx264 -t 10 -pix_fmt yuv420p -movflags +faststart -y "$output"
}

# Generate all error videos based on Zurg error messages
generate_error_video "All tokens are expired\\nPlease wait for token refresh" "token_expired.mp4"
generate_error_video "Download quota exceeded\\nPlease wait or upgrade account" "quota_exceeded.mp4"
generate_error_video "Download link has expired\\nZurg is repairing this file" "expired_link.mp4"
generate_error_video "Failed to generate download link\\nTry again later" "generation_failed.mp4"
generate_error_video "Traffic limit reached\\nPlease wait" "traffic_limit.mp4"
generate_error_video "Network connection error\\nCheck internet connection" "network_error.mp4"
generate_error_video "Server error occurred\\nPlease try again" "server_error.mp4"
generate_error_video "Request timed out\\nPlease try again" "timeout_error.mp4"
generate_error_video "Cannot unrestrict file\\nZurg is repairing it" "cannot_unrestrict_file.mp4"

# Generate fallback/generic error video
generate_error_video "File temporarily unavailable\\nZurg is working to fix this" "generic_error.mp4"

echo "All error videos generated in: $VIDEO_DIR"
echo ""
echo "Generated files:"
ls -la "$VIDEO_DIR"/*.mp4 2>/dev/null || echo "No MP4 files found"