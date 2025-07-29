# Caddy Zurg Error Handler Plugin

Caddy plugin that intercepts HTTP 500 errors from Zurg STRM endpoints and serves appropriate error videos based on the specific error messages.

## Features

- intercepts 500 errors from `/strm/*` endpoints
- matches specific error messages to appropriate error videos
- serves custom error videos instead of generic HTTP errors
- configurable error message patterns and video mappings
- automatic fallback to generic error video

## Caddy binary

To run Caddy as a standalone binary (without Docker):

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
git clone https://github.com/andesco/caddy-plugin-zurg.git
cd caddy-plugin-zurg
xcaddy build --with . # creates `caddy` binary in current directory
```

## Docker Compose

To run Caddy within Docker:

### 1.  **Create `Dockerfile.caddy`:**

```Dockerfile
FROM caddy:builder AS builder

RUN xcaddy build \
    --with github.com/andesco/caddy-plugin-zurg

FROM caddy:latest

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
```

### 2.  **Update `docker-compose.yml`:**
   
Modify your `docker-compose.yml` to build a new Caddy image that combines the newly-built `caddy` binary to the existing `caddy:latest` image:

```yaml
    caddy:
      #image: caddy:latest
      build:
        context: . # current directory (location of Dockerfile.caddy)
        dockerfile: Dockerfile.caddy
      container_name: caddy
```

### 3. **Update your Caddyfile configuration:**


```caddy
    zurg.example.com {
        zurg_error_handler
        reverse_proxy localhost:9999
    }
```

```caddy
    zurg.example.com {
        zurg_error_handler {
            video_path "/etc/zurg/zurg.git/error_videos" # optional
            strm_paths "/strm/" # optional
        }
        reverse_proxy localhost:9999
    }
```

    
### 4.  **Build and start your services:**
```bash
    docker-compose up --build -d
```

### 5. Generate Error Videos

The plugin relies on video files to display errors. Run the included script to generate videos for all common Zurg error scenarios in the `error_videos` directory.

```bash
./generate_error_videos.sh
```

The script uses `ffmpeg` to generate each video, for example:

```bash
read -p "error message to display via streaming video:" error_message \
  && error_message=${error_message:-media file not found}
ffmpeg -f lavfi -i color=c=black:s=1920x1080:d=10 \
  -vf "drawtext=fontfile=/path/to/font.ttf:text='${error_message}':fontsize=96:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2" \
  -c:v libx264 -t 10 -pix_fmt yuv420p -movflags +faststart not_found.mp4
```

## Error Message Mappings

The plugin automatically maps these Zurg error messages to videos:

| Error Message Pattern | Video File |
|----------------------|------------|
| `all tokens are expired` | `token_expired.mp4` |
| `bytes_limit_reached` | `quota_exceeded.mp4` |
| `invalid_download_code` | `expired_link.mp4` |
| `failed_generation` | `generation_failed.mp4` |
| `traffic_exhausted` | `traffic_limit.mp4` |
| `unrestrict link request failed` | `network_error.mp4` |
| `unreadable body` | `server_error.mp4` |
| `timeout` | `timeout_error.mp4` |
| `connection reset` | `network_error.mp4` |
| default | `cannot_unrestrict_file.mp4` |

## How It Works

1. **Request Interception**: plugin intercepts requests to `/strm/*` paths
2. **Response Monitoring**: captures responses from Zurg
3. **Error Detection**: identifies HTTP 500 responses
4. **Content Analysis**: reads response body to find specific error messages
5. **Video Mapping**: matches error patterns to appropriate video files
6. **Redirect**: issues HTTP 302 redirect to the error video

Test the plugin by triggering an STRM error:

```bash
curl -v http://zurg.example.com/strm/invalid_code
```

## Development

To modify error mappings or add new patterns:

1. update `ErrorMappings` in `plugin.go`
2. add corresponding video generation in `generate_error_videos.sh`
3. rebuild Caddy with updated plugin

