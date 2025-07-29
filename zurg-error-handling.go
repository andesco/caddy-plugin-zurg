package zurg_error_handling

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ZurgErrorHandler{})
	httpcaddyfile.RegisterHandlerDirective("zurg_error_handler", parseCaddyfile)
}

// ZurgErrorHandler intercepts 500 errors from Zurg STRM endpoints and serves appropriate error videos
type ZurgErrorHandler struct {
	// VideoPath is the directory containing error videos
	VideoPath string `json:"video_path,omitempty"`
	
	// ErrorMappings maps error message patterns to video filenames
	ErrorMappings map[string]string `json:"error_mappings,omitempty"`
	
	// STRMPaths defines which paths this handler should intercept
	STRMPaths []string `json:"strm_paths,omitempty"`
	
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information
func (ZurgErrorHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.zurg_error_handler",
		New: func() caddy.Module { return new(ZurgErrorHandler) },
	}
}

// Provision implements caddy.Provisioner
func (z *ZurgErrorHandler) Provision(ctx caddy.Context) error {
	z.logger = ctx.Logger(z)
	
	// Set default error mappings if none provided
	if z.ErrorMappings == nil {
		z.ErrorMappings = map[string]string{
			"all tokens are expired":           "token_expired.mp4",
			"bytes_limit_reached":              "quota_exceeded.mp4",
			"invalid_download_code":            "expired_link.mp4",
			"failed_generation":                "generation_failed.mp4",
			"traffic_exhausted":                "traffic_limit.mp4",
			"unrestrict link request failed":   "network_error.mp4",
			"unreadable body":                  "server_error.mp4",
			"undecodable response":             "server_error.mp4",
			"timeout":                          "timeout_error.mp4",
			"connection reset by peer":         "network_error.mp4",
			"EOF":                              "network_error.mp4",
			"broken pipe":                      "network_error.mp4",
		}
	}
	
	// Set default STRM paths if none provided
	if z.STRMPaths == nil {
		z.STRMPaths = []string{"/strm/"}
	}
	
	// Set default video path if none provided
	if z.VideoPath == "" {
		z.VideoPath = "/etc/zurg/zurg.git/error_videos"
	}
	
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler
func (z ZurgErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Check if this is a STRM path we should intercept
	shouldIntercept := false
	for _, path := range z.STRMPaths {
		if strings.HasPrefix(r.URL.Path, path) {
			shouldIntercept = true
			break
		}
	}
	
	if !shouldIntercept {
		return next.ServeHTTP(w, r)
	}
	
	// Create a response recorder to capture the response
	rec := caddyhttp.NewResponseRecorder(w, nil, nil)
	
	// Call the next handler
	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	
	// Check if this is a 500 error
	if rec.Status() == http.StatusInternalServerError {
		// Read the response body to check for error messages
		body := rec.Buffer().Bytes()
		bodyStr := string(body)
		
		z.logger.Info("Intercepted 500 error from STRM endpoint",
			zap.String("path", r.URL.Path),
			zap.String("error_body", bodyStr))
		
		// Find matching error message and redirect to appropriate video
		videoFile := z.findMatchingVideo(bodyStr)
		if videoFile != "" {
			videoURL := fmt.Sprintf("%s/%s", z.VideoPath, videoFile)
			z.logger.Info("Redirecting to error video",
				zap.String("video", videoFile),
				zap.String("url", videoURL))
			
			// Redirect to the error video
			http.Redirect(w, r, videoURL, http.StatusTemporaryRedirect)
			return nil
		}
	}
	
	// Copy the recorded response back to the original writer
	w.WriteHeader(rec.Status())
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	
	// Copy response body
	if rec.Buffer() != nil {
		_, err := io.Copy(w, bytes.NewReader(rec.Buffer().Bytes()))
		if err != nil {
			return err
		}
	}
	
	return nil
}

// findMatchingVideo finds the appropriate video file based on error message content
func (z *ZurgErrorHandler) findMatchingVideo(errorBody string) string {
	errorBody = strings.ToLower(errorBody)
	
	// Check for specific error patterns
	for pattern, videoFile := range z.ErrorMappings {
		if strings.Contains(errorBody, strings.ToLower(pattern)) {
			return videoFile
		}
	}
	
	// Default fallback video for any STRM error
	if strings.Contains(errorBody, "failed to unrestrict link") {
		return "cannot_unrestrict_file.mp4"
	}
	
	return ""
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var z ZurgErrorHandler
	err := z.UnmarshalCaddyfile(h.Dispenser)
	return z, err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (z *ZurgErrorHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return d.Err("expected token following filter")
	}

	for d.NextBlock(0) {
		switch d.Val() {
		case "video_path":
			if !d.Args(&z.VideoPath) {
				return d.ArgErr()
			}
		case "strm_paths":
			z.STRMPaths = d.RemainingArgs()
		case "error_mapping":
			if !d.NextArg() {
				return d.ArgErr()
			}
			pattern := d.Val()
			if !d.NextArg() {
				return d.ArgErr()
			}
			videoFile := d.Val()
			
			if z.ErrorMappings == nil {
				z.ErrorMappings = make(map[string]string)
			}
			z.ErrorMappings[pattern] = videoFile
		}
	}
	
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ZurgErrorHandler)(nil)
	_ caddyhttp.MiddlewareHandler = (*ZurgErrorHandler)(nil)
	_ caddyfile.Unmarshaler       = (*ZurgErrorHandler)(nil)
)