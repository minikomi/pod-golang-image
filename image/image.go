package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"os"

	transit "github.com/babashka/transit-go"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// formatToMediaType converts image format strings to MIME types
func formatToMediaType(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "image/" + format
	}
}

// Info returns image dimensions and media type by reading only the header
func Info(path string) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	return map[transit.Keyword]interface{}{
		transit.Keyword("width"):      config.Width,
		transit.Keyword("height"):     config.Height,
		transit.Keyword("media-type"): formatToMediaType(format),
	}, nil
}

// ToBase64 encodes an image as base64 without resizing
func ToBase64(path string) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Re-encode the image
	var buf bytes.Buffer
	var outputFormat string

	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		outputFormat = "jpeg"
	default:
		err = png.Encode(&buf, img)
		outputFormat = "png"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	return map[transit.Keyword]interface{}{
		transit.Keyword("data"):       encoded,
		transit.Keyword("width"):      width,
		transit.Keyword("height"):     height,
		transit.Keyword("media-type"): formatToMediaType(outputFormat),
	}, nil
}

// ResizeOptions contains options for resizing
type ResizeOptions struct {
	MaxEdge   int
	MaxWidth  int
	MaxHeight int
	Format    string
	Quality   int
}

// parseResizeOptions extracts resize options from the transit map
func parseResizeOptions(opts map[interface{}]interface{}) ResizeOptions {
	result := ResizeOptions{
		Quality: 85, // default quality
	}

	for k, v := range opts {
		key, ok := k.(transit.Keyword)
		if !ok {
			continue
		}

		switch string(key) {
		case "max-edge":
			if val, ok := v.(int64); ok {
				result.MaxEdge = int(val)
			} else if val, ok := v.(int); ok {
				result.MaxEdge = val
			}
		case "max-width":
			if val, ok := v.(int64); ok {
				result.MaxWidth = int(val)
			} else if val, ok := v.(int); ok {
				result.MaxWidth = val
			}
		case "max-height":
			if val, ok := v.(int64); ok {
				result.MaxHeight = int(val)
			} else if val, ok := v.(int); ok {
				result.MaxHeight = val
			}
		case "format":
			if val, ok := v.(string); ok {
				result.Format = val
			}
		case "quality":
			if val, ok := v.(int64); ok {
				result.Quality = int(val)
			} else if val, ok := v.(int); ok {
				result.Quality = val
			}
		}
	}

	return result
}

// computeTargetSize calculates the target dimensions based on constraints
func computeTargetSize(srcWidth, srcHeight int, opts ResizeOptions) (int, int) {
	if opts.MaxEdge == 0 && opts.MaxWidth == 0 && opts.MaxHeight == 0 {
		// No constraints, return original size
		return srcWidth, srcHeight
	}

	scale := 1.0
	needsResize := false

	// Check max-edge constraint
	if opts.MaxEdge > 0 {
		maxDim := srcWidth
		if srcHeight > maxDim {
			maxDim = srcHeight
		}
		if maxDim > opts.MaxEdge {
			edgeScale := float64(opts.MaxEdge) / float64(maxDim)
			if !needsResize || edgeScale < scale {
				scale = edgeScale
				needsResize = true
			}
		}
	}

	// Check max-width constraint
	if opts.MaxWidth > 0 && srcWidth > opts.MaxWidth {
		widthScale := float64(opts.MaxWidth) / float64(srcWidth)
		if !needsResize || widthScale < scale {
			scale = widthScale
			needsResize = true
		}
	}

	// Check max-height constraint
	if opts.MaxHeight > 0 && srcHeight > opts.MaxHeight {
		heightScale := float64(opts.MaxHeight) / float64(srcHeight)
		if !needsResize || heightScale < scale {
			scale = heightScale
			needsResize = true
		}
	}

	if !needsResize {
		// Image already fits within constraints
		return srcWidth, srcHeight
	}

	// Calculate target dimensions
	targetWidth := int(float64(srcWidth) * scale)
	targetHeight := int(float64(srcHeight) * scale)

	// Ensure at least 1x1
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}

	return targetWidth, targetHeight
}

// Resize resizes an image according to the provided options
func Resize(path string, opts ResizeOptions) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Determine target size
	targetWidth, targetHeight := computeTargetSize(srcWidth, srcHeight, opts)

	// Determine output format
	outputFormat := opts.Format
	if outputFormat == "" {
		// Use input format, but only support jpeg and png for output
		if format == "jpeg" {
			outputFormat = "jpeg"
		} else {
			outputFormat = "png"
		}
	}

	// Resize if needed
	var resultImg image.Image
	if targetWidth != srcWidth || targetHeight != srcHeight {
		// Create destination image
		dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
		// Use CatmullRom (high quality) scaling
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		resultImg = dst
	} else {
		resultImg = img
	}

	// Encode to requested format
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		quality := opts.Quality
		if quality < 1 || quality > 100 {
			quality = 85
		}
		err = jpeg.Encode(&buf, resultImg, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, resultImg)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	return map[transit.Keyword]interface{}{
		transit.Keyword("data"):       encoded,
		transit.Keyword("width"):      targetWidth,
		transit.Keyword("height"):     targetHeight,
		transit.Keyword("media-type"): formatToMediaType(outputFormat),
	}, nil
}
