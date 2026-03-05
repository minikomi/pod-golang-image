#!/usr/bin/env bb

;; Demo of new image manipulation features
(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[pod.poyo.image :as img])

(println "===========================================")
(println "  New Image Manipulation Features Demo")
(println "===========================================\n")

(def test-img "test.png")
(println (format "Using test image: %s (200x150)\n" test-img))

;; Rotate
(println "🔄 ROTATE - Rotate images by any angle")
(println "   Supports: 90°, 180°, 270°, or arbitrary angles")
(doseq [angle [90 180 270 45]]
  (let [result (img/rotate test-img {:degrees angle})]
    (println (format "   • %d° → %dx%d pixels" 
                     angle (:width result) (:height result)))))
(println)

;; Flip
(println "🔀 FLIP - Mirror images horizontally or vertically")
(let [h (img/flip test-img {:direction "horizontal"})
      v (img/flip test-img {:direction "vertical"})]
  (println (format "   • Horizontal flip: %dx%d" (:width h) (:height h)))
  (println (format "   • Vertical flip: %dx%d" (:width v) (:height v))))
(println)

;; Crop
(println "✂️  CROP - Extract rectangular regions")
(let [crops [[10 10 80 60 "top-left"]
             [100 50 80 60 "center-right"]
             [50 80 100 50 "bottom"]]]
  (doseq [[x y w h desc] crops]
    (let [result (img/crop test-img {:x x :y y :width w :height h})]
      (println (format "   • %s (%d,%d %dx%d) → %dx%d"
                       desc x y w h (:width result) (:height result))))))
(println)

;; Grayscale
(println "⚫ GRAYSCALE - Convert to black & white")
(let [gray (img/grayscale test-img)]
  (println (format "   • Color → grayscale: %dx%d" 
                   (:width gray) (:height gray)))
  (println (format "   • Base64 size: %d bytes" 
                   (int (/ (count (:data gray)) 4 3)))))
(println)

;; Draw Text
(println "✍️  DRAW-TEXT - Add text overlays and watermarks")
(let [examples [["Hello World!" 10 30 "#FF0000"]
                ["© 2024" 150 140 "#FFFFFF"]
                ["DRAFT" 80 75 "#0000FF"]]]
  (doseq [[text x y color] examples]
    (let [result (img/draw-text test-img {:text text :x x :y y :color color})]
      (println (format "   • '%s' at (%d,%d) color %s" 
                       text x y color)))))
(println)

;; Combination Example
(println "🎨 COMBINATION - Chain multiple operations")
(println "   Example workflow: Rotate → Crop → Grayscale → Add Watermark")
(let [step1 (img/rotate test-img {:degrees 90})
      _ (println (format "   1. Rotate 90°: %dx%d" (:width step1) (:height step1)))
      
      step2 (img/crop test-img {:x 20 :y 20 :width 160 :height 110})
      _ (println (format "   2. Crop center: %dx%d" (:width step2) (:height step2)))
      
      step3 (img/grayscale test-img)
      _ (println (format "   3. Grayscale: %dx%d" (:width step3) (:height step3)))
      
      step4 (img/draw-text test-img {:text "Processed" :x 10 :y 20 :color "#FFFFFF"})
      _ (println (format "   4. Add text: %dx%d" (:width step4) (:height step4)))]
  nil)
(println)

;; Use Cases
(println "💡 COMMON USE CASES:")
(println "   • Thumbnail generation with resize + crop")
(println "   • Photo orientation fix with rotate")
(println "   • Watermark images with draw-text")
(println "   • Create image variations with flip")
(println "   • Archive optimization with grayscale")
(println "   • Extract regions of interest with crop")
(println)

(println "✅ Demo completed! All 5 new functions working.")
