# Caddy Zurg Error Handler Plugin

Caddy plugin that intercepts HTTP 500 errors from Zurg STRM endpoints and serves appropriate error videos based on the specific error messages.

## Features

- intercepts 500 errors from `/strm/*` endpoints
- matches specific error messages to appropriate error videos
- serves custom error videos instead of generic HTTP errors
- configurable error message patterns and video mappings
- automatic fallback to generic error video

## Caddy binary

To build and run Caddy as a standalone binary (without Docker):

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
mkdir -p /etc/zurg
git clone https://github.com/andesco/caddy-plugin-zurg
cd /etc/zurg/caddy-plugin-zurg
xcaddy build --with . # creates `caddy` binary in current directory
```

## Docker Compose

To build and run Caddy within a Docker container

### 1. Download `andesco/caddy-plugin-zurg`
```bash
cd /etc/zurg/
git clone https://github.com/andesco/caddy-plugin-zurg
```

### 2.  Update `docker-compose.yml`
   
Modify your `docker-compose.yml` to:
1. build a new Caddy binary that includes this plugin within a `caddy:builder` image; and
2. mount the `Caddyfile` and `error_videos` from the cloned repository.

```yaml
    caddy:
      #image: caddy:latest
      build:
        context: /etc/zurg/caddy-plugin-zurg/ # path for Dockerfile.caddy
        dockerfile: Dockerfile.caddy
      container_name: caddy
      volumes:
        - /etc/zurg/caddy-plugin-zurg/Caddyfile:/etc/caddy/Caddyfile:ro
        - /etc/zurg/caddy-plugin-zurg/error_videos:/etc/zurg/caddy-plugin-zurg/error_videos:ro
        - caddy_data:/data
        - caddy_config:/config
```

### 3. Update `Caddyfile`

Update `caddy-plugin-zurg/Caddyfile` with your domain/hostname and any paths that differ from these defaults:

```caddy
{
    order zurg_error_handler before reverse_proxy
}

zurg.example.com {
    zurg_error_handler {
        video_path "/etc/zurg/caddy-plugin-zurg/error_videos" # optional
        strm_paths "/strm/" # optional
    }
    reverse_proxy zurg:9999
}
```
    
### 4. Build and start all services:

```bash
docker compose up --build -d
```

### 5. Error Videos

The necessary error videos are included directly in the `error_videos` directory within this repository. No generation is required for basic setup.

If you wish to regenerate or customize the error videos, you will need `ffmpeg` installed on your system. The `generate_error_videos.sh` script can be used for this purpose.

```bash
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

