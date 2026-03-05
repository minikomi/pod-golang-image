# pod-golang-image - Complete Feature List

## Overview

A comprehensive babashka pod for image manipulation in Go, exposing 8 functions across multiple categories.

## Functions by Category

### 📊 Information
- **`info`** - Get image dimensions and MIME type (header-only, very fast)

### 🔄 Transformations
- **`resize`** - High-quality resizing with Catmull-Rom interpolation
  - Constraints: max-edge, max-width, max-height
  - Format conversion (PNG/JPEG)
  - Quality control
  - Aspect ratio preservation

### 🎨 Geometric Operations
- **`rotate`** - Rotate by any angle
  - Efficient 90°, 180°, 270° rotations
  - Arbitrary angle support (45°, 33.5°, etc.)
  - Automatic bounds calculation
  
- **`flip`** - Mirror images
  - Horizontal flip
  - Vertical flip
  
- **`crop`** - Extract rectangular regions
  - Precise pixel-level control
  - Automatic bounds clamping

### 🎨 Color Operations
- **`grayscale`** - Convert to black & white
  - Preserves image dimensions
  - Automatic color space conversion

### ✍️ Drawing & Text
- **`draw-text`** - Add text overlays
  - Watermarking
  - Custom positioning
  - Hex color support (#RRGGBB)
  - Named colors (black, white, red, green, blue)
  - Uses built-in 7x13 bitmap font

### 💾 Encoding
- **`to-base64`** - Encode images as base64
  - No resizing
  - Format detection
  - Ready for embedding

## Format Support

| Operation | JPEG | PNG | GIF | WebP |
|-----------|------|-----|-----|------|
| Decode    | ✅   | ✅  | ✅  | ✅   |
| Encode    | ✅   | ✅  | ❌  | ❌   |

## Performance Characteristics

| Function     | Speed      | Memory | Notes |
|--------------|------------|--------|-------|
| info         | Very Fast  | Low    | Header-only, no full decode |
| resize       | Fast       | Medium | Catmull-Rom = high quality |
| rotate 90°   | Fast       | Medium | Optimized pixel copy |
| rotate 45°   | Moderate   | Medium | Needs interpolation |
| flip         | Very Fast  | Medium | Simple pixel remap |
| crop         | Very Fast  | Low    | Direct pixel copy |
| grayscale    | Fast       | Medium | Color space conversion |
| draw-text    | Fast       | Medium | Bitmap font rendering |
| to-base64    | Fast       | Medium | Re-encode only |

## Use Cases

### 🖼️ Photo Processing
```clojure
;; Fix orientation
(img/rotate "photo.jpg" {:degrees 90})

;; Create thumbnail
(img/resize "photo.jpg" {:max-edge 200})

;; Add watermark
(img/draw-text "photo.jpg" {:text "© 2024" :x 10 :y 30 :color "#FFFFFF"})
```

### 📱 Web Optimization
```clojure
;; Responsive images
(img/resize "hero.jpg" {:max-width 1920})
(img/resize "hero.jpg" {:max-width 768})

;; Convert to smaller format
(img/resize "photo.png" {:max-edge 1200 :format "jpeg" :quality 85})
```

### 🎨 Image Processing Pipelines
```clojure
;; Crop → Grayscale → Add text
(-> (img/crop "photo.jpg" {:x 100 :y 100 :width 800 :height 600})
    (:data)
    ;; Would need to save/reload for chaining
    )
```

### 📄 Document Processing
```clojure
;; Extract regions
(img/crop "scan.png" {:x 50 :y 50 :width 700 :height 900})

;; Convert to grayscale for storage
(img/grayscale "scan.png")
```

### 🔄 Batch Operations
```clojure
;; Process all images in a directory
(doseq [file (file-seq (io/file "photos"))]
  (when (.endsWith (.getName file) ".jpg")
    (let [thumb (img/resize (.getPath file) {:max-edge 200})]
      (spit (str "thumbs/" (.getName file)) (:data thumb)))))
```

## Implementation Quality

- ✅ All functions return consistent transit+json maps
- ✅ Proper error handling with descriptive messages
- ✅ Type checking for all arguments
- ✅ Bounds checking and clamping
- ✅ Efficient algorithms (no unnecessary copies)
- ✅ Base64 encoding for easy transport
- ✅ Comprehensive test coverage

## Code Statistics

- **Total Lines:** 1,213
- **Functions:** 8 exposed + 4 internal helpers
- **Binary Size:** 4.4 MB
- **Dependencies:** 3 (bencode-go, transit-go, x/image)
- **Test Coverage:** 8/8 functions tested

## Future Enhancements

Potential additions (not yet implemented):
- Blur (Gaussian, box)
- Sharpen (unsharp mask)
- Brightness/Contrast adjustment
- Saturation control
- Animated GIF creation
- Custom fonts for text rendering
- More drawing primitives (rectangles, circles)
- Image composition/overlay
- Color palette manipulation
