# pod-golang-image

A [babashka pod](https://github.com/babashka/pods) that exposes Go's image library for image manipulation.

## Features

- **Fast image info** - Get dimensions and MIME type without loading the full image
- **High-quality resizing** - Uses Catmull-Rom interpolation for excellent quality
- **Multiple formats** - Decode JPEG, PNG, GIF, WebP; encode JPEG, PNG
- **Base64 encoding** - Return images as base64 for easy embedding

## Installation

Build from source:

```bash
go build -o pod-golang-image .
```

## Usage

```clojure
#!/usr/bin/env bb
(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[pod.poyo.image :as img])

;; Get image info (fast, header-only)
(img/info "/path/to/photo.jpg")
;; => {:width 4000 :height 3000 :media-type "image/jpeg"}

;; Resize with max edge constraint
(img/resize "/path/to/photo.jpg" {:max-edge 1024})
;; => {:data "iVBOR..." :width 1024 :height 768 :media-type "image/png"}

;; Resize with max width
(img/resize "/path/to/photo.jpg" {:max-width 800})
;; => {:data "iVBOR..." :width 800 :height 600 :media-type "image/jpeg"}

;; Resize with format and quality options
(img/resize "/path/to/photo.jpg" {:max-height 600 :format "png" :quality 90})
;; => {:data "iVBOR..." :width 800 :height 600 :media-type "image/png"}

;; Encode to base64 without resizing
(img/to-base64 "/path/to/photo.jpg")
;; => {:data "..." :width 4000 :height 3000 :media-type "image/jpeg"}
```

## API

### `pod.poyo.image/info`

Get image dimensions and MIME type without loading the full image.

**Arguments:**
- `path` - String path to image file

**Returns:**
```clojure
{:width 4000
 :height 3000
 :media-type "image/jpeg"}
```

### `pod.poyo.image/resize`

Resize an image and return as base64.

**Arguments:**
- `path` - String path to image file
- `opts` - Optional map of options:
  - `:max-edge N` - Scale so max(width, height) <= N, preserve aspect ratio
  - `:max-width N` - Scale so width <= N, preserve aspect ratio
  - `:max-height N` - Scale so height <= N, preserve aspect ratio
  - `:format "png"` or `"jpeg"` - Output format (default: same as input)
  - `:quality 85` - JPEG quality 1-100 (default 85)

If multiple constraints are given, the most restrictive is used.
If the image already fits within constraints, it won't be upscaled.

**Returns:**
```clojure
{:data "iVBOR...base64..."  ; base64-encoded image
 :width 1024
 :height 768
 :media-type "image/png"}
```

### `pod.poyo.image/to-base64`

Encode an image as base64 without resizing.

**Arguments:**
- `path` - String path to image file

**Returns:**
```clojure
{:data "...base64..."
 :width 4000
 :height 3000
 :media-type "image/jpeg"}
```

## Format Support

**Decode (input):** JPEG, PNG, GIF, WebP  
**Encode (output):** JPEG, PNG

## License

MIT
