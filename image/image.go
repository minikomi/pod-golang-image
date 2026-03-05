package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"

	transit "github.com/babashka/transit-go"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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

// Rotate rotates an image by the specified degrees (90, 180, 270, or arbitrary angle)
func Rotate(path string, opts map[interface{}]interface{}) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Parse rotation angle
	degrees := 0.0
	if val, ok := opts[transit.Keyword("degrees")]; ok {
		if d, ok := val.(int64); ok {
			degrees = float64(d)
		} else if d, ok := val.(float64); ok {
			degrees = d
		} else if d, ok := val.(int); ok {
			degrees = float64(d)
		}
	}

	var resultImg image.Image

	// Handle common rotations efficiently
	switch int(degrees) % 360 {
	case 0:
		resultImg = img
	case 90, -270:
		resultImg = rotate90(img)
	case 180, -180:
		resultImg = rotate180(img)
	case 270, -90:
		resultImg = rotate270(img)
	default:
		// Arbitrary rotation (more expensive)
		resultImg = rotateArbitrary(img, degrees)
	}

	// Determine output format
	outputFormat := format
	if format != "jpeg" && format != "png" {
		outputFormat = "png"
	}

	// Encode result
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, resultImg, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, resultImg)
		outputFormat = "png"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	bounds := resultImg.Bounds()

	return map[transit.Keyword]interface{}{
		transit.Keyword("data"):       encoded,
		transit.Keyword("width"):      bounds.Dx(),
		transit.Keyword("height"):     bounds.Dy(),
		transit.Keyword("media-type"): formatToMediaType(outputFormat),
	}, nil
}

// rotate90 rotates an image 90 degrees clockwise
func rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(height-1-y, x, img.At(x, y))
		}
	}
	return rotated
}

// rotate180 rotates an image 180 degrees
func rotate180(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(width-1-x, height-1-y, img.At(x, y))
		}
	}
	return rotated
}

// rotate270 rotates an image 270 degrees clockwise (90 counter-clockwise)
func rotate270(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	rotated := image.NewRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(y, width-1-x, img.At(x, y))
		}
	}
	return rotated
}

// rotateArbitrary rotates an image by an arbitrary angle (in degrees)
func rotateArbitrary(img image.Image, degrees float64) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert to radians
	rad := degrees * math.Pi / 180.0
	cos, sin := math.Cos(rad), math.Sin(rad)

	// Calculate new bounds
	w, h := float64(width), float64(height)
	newWidth := int(math.Abs(w*cos) + math.Abs(h*sin))
	newHeight := int(math.Abs(w*sin) + math.Abs(h*cos))

	rotated := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Center points
	cx, cy := float64(width)/2, float64(height)/2
	newCx, newCy := float64(newWidth)/2, float64(newHeight)/2

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Reverse rotation to find source pixel
			dx, dy := float64(x)-newCx, float64(y)-newCy
			srcX := int(dx*cos+dy*sin + cx)
			srcY := int(-dx*sin+dy*cos + cy)

			if srcX >= 0 && srcX < width && srcY >= 0 && srcY < height {
				rotated.Set(x, y, img.At(srcX, srcY))
			}
		}
	}
	return rotated
}

// Flip flips an image horizontally or vertically
func Flip(path string, opts map[interface{}]interface{}) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Parse direction
	direction := "horizontal"
	if val, ok := opts[transit.Keyword("direction")]; ok {
		if s, ok := val.(string); ok {
			direction = s
		}
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	flipped := image.NewRGBA(image.Rect(0, 0, width, height))

	if direction == "horizontal" {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				flipped.Set(width-1-x, y, img.At(x, y))
			}
		}
	} else {
		// vertical
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				flipped.Set(x, height-1-y, img.At(x, y))
			}
		}
	}

	// Determine output format
	outputFormat := format
	if format != "jpeg" && format != "png" {
		outputFormat = "png"
	}

	// Encode result
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, flipped, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, flipped)
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

// Crop extracts a rectangular region from an image
func Crop(path string, opts map[interface{}]interface{}) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Parse crop rectangle
	x, y, width, height := 0, 0, 0, 0
	if val, ok := opts[transit.Keyword("x")]; ok {
		if i, ok := val.(int64); ok {
			x = int(i)
		} else if i, ok := val.(int); ok {
			x = i
		}
	}
	if val, ok := opts[transit.Keyword("y")]; ok {
		if i, ok := val.(int64); ok {
			y = int(i)
		} else if i, ok := val.(int); ok {
			y = i
		}
	}
	if val, ok := opts[transit.Keyword("width")]; ok {
		if i, ok := val.(int64); ok {
			width = int(i)
		} else if i, ok := val.(int); ok {
			width = i
		}
	}
	if val, ok := opts[transit.Keyword("height")]; ok {
		if i, ok := val.(int64); ok {
			height = int(i)
		} else if i, ok := val.(int); ok {
			height = i
		}
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("width and height must be positive")
	}

	// Clamp to image bounds
	bounds := img.Bounds()
	imgWidth, imgHeight := bounds.Dx(), bounds.Dy()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+width > imgWidth {
		width = imgWidth - x
	}
	if y+height > imgHeight {
		height = imgHeight - y
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("crop region is outside image bounds")
	}

	// Create cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, width, height))
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			cropped.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}

	// Determine output format
	outputFormat := format
	if format != "jpeg" && format != "png" {
		outputFormat = "png"
	}

	// Encode result
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, cropped, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, cropped)
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

// Grayscale converts an image to grayscale
func Grayscale(path string) (map[transit.Keyword]interface{}, error) {
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
	width, height := bounds.Dx(), bounds.Dy()
	gray := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	// Determine output format
	outputFormat := format
	if format != "jpeg" && format != "png" {
		outputFormat = "png"
	}

	// Encode result
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, gray, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, gray)
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

// parseColor parses a color string (hex format like "#FF0000" or named colors)
func parseColor(colorStr string) (color.Color, error) {
	if colorStr == "" {
		return color.Black, nil
	}

	// Handle hex colors
	if colorStr[0] == '#' {
		colorStr = colorStr[1:]
		if len(colorStr) == 6 {
			var r, g, b uint8
			fmt.Sscanf(colorStr, "%02x%02x%02x", &r, &g, &b)
			return color.RGBA{r, g, b, 255}, nil
		}
	}

	// Named colors
	switch colorStr {
	case "black":
		return color.Black, nil
	case "white":
		return color.White, nil
	case "red":
		return color.RGBA{255, 0, 0, 255}, nil
	case "green":
		return color.RGBA{0, 255, 0, 255}, nil
	case "blue":
		return color.RGBA{0, 0, 255, 255}, nil
	default:
		return color.Black, nil
	}
}

// DrawText draws text on an image
func DrawText(path string, opts map[interface{}]interface{}) (map[transit.Keyword]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Parse options
	text := ""
	x, y := 10, 20
	colorStr := "#000000"

	if val, ok := opts[transit.Keyword("text")]; ok {
		if s, ok := val.(string); ok {
			text = s
		}
	}
	if val, ok := opts[transit.Keyword("x")]; ok {
		if i, ok := val.(int64); ok {
			x = int(i)
		} else if i, ok := val.(int); ok {
			x = i
		}
	}
	if val, ok := opts[transit.Keyword("y")]; ok {
		if i, ok := val.(int64); ok {
			y = int(i)
		} else if i, ok := val.(int); ok {
			y = i
		}
	}
	if val, ok := opts[transit.Keyword("color")]; ok {
		if s, ok := val.(string); ok {
			colorStr = s
		}
	}

	textColor, err := parseColor(colorStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse color: %w", err)
	}

	// Convert to RGBA for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Draw text using basic font
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)

	// Determine output format
	outputFormat := format
	if format != "jpeg" && format != "png" {
		outputFormat = "png"
	}

	// Encode result
	var buf bytes.Buffer
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 85})
	default:
		err = png.Encode(&buf, rgba)
		outputFormat = "png"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	width, height := bounds.Dx(), bounds.Dy()

	return map[transit.Keyword]interface{}{
		transit.Keyword("data"):       encoded,
		transit.Keyword("width"):      width,
		transit.Keyword("height"):     height,
		transit.Keyword("media-type"): formatToMediaType(outputFormat),
	}, nil
}
