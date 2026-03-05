#!/usr/bin/env bb

;; Test new image manipulation functions
(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[co.poyo.pod-golang-image :as img])

(println "Testing New Image Functions")
(println "=")

;; Test 1: Rotate
(println "\n1. Testing rotate:")
(let [result-90 (img/rotate "test.png" {:degrees 90})
      result-180 (img/rotate "test.png" {:degrees 180})
      result-270 (img/rotate "test.png" {:degrees 270})]
  (println (format "   90° rotation: %dx%d (was 200x150)" 
                   (:width result-90) (:height result-90)))
  (assert (= 150 (:width result-90)) "90° should swap dimensions")
  (assert (= 200 (:height result-90)) "90° should swap dimensions")
  
  (println (format "   180° rotation: %dx%d (was 200x150)" 
                   (:width result-180) (:height result-180)))
  (assert (= 200 (:width result-180)) "180° should keep dimensions")
  (assert (= 150 (:height result-180)) "180° should keep dimensions")
  
  (println (format "   270° rotation: %dx%d (was 200x150)" 
                   (:width result-270) (:height result-270)))
  (assert (= 150 (:width result-270)) "270° should swap dimensions")
  (assert (= 200 (:height result-270)) "270° should swap dimensions")
  (println "   ✓ rotate test passed!"))

;; Test 2: Flip
(println "\n2. Testing flip:")
(let [h-flip (img/flip "test.png" {:direction "horizontal"})
      v-flip (img/flip "test.png" {:direction "vertical"})]
  (println (format "   Horizontal flip: %dx%d" (:width h-flip) (:height h-flip)))
  (assert (= 200 (:width h-flip)) "Flip should preserve dimensions")
  (assert (= 150 (:height h-flip)) "Flip should preserve dimensions")
  
  (println (format "   Vertical flip: %dx%d" (:width v-flip) (:height v-flip)))
  (assert (= 200 (:width v-flip)) "Flip should preserve dimensions")
  (assert (= 150 (:height v-flip)) "Flip should preserve dimensions")
  (println "   ✓ flip test passed!"))

;; Test 3: Crop
(println "\n3. Testing crop:")
(let [cropped (img/crop "test.png" {:x 50 :y 40 :width 100 :height 60})]
  (println (format "   Cropped from (50,40) with size 100x60: %dx%d" 
                   (:width cropped) (:height cropped)))
  (assert (= 100 (:width cropped)) "Crop should have requested width")
  (assert (= 60 (:height cropped)) "Crop should have requested height")
  (println "   ✓ crop test passed!"))

;; Test 4: Grayscale
(println "\n4. Testing grayscale:")
(let [gray (img/grayscale "test.png")]
  (println (format "   Grayscale: %dx%d" (:width gray) (:height gray)))
  (assert (= 200 (:width gray)) "Grayscale should preserve dimensions")
  (assert (= 150 (:height gray)) "Grayscale should preserve dimensions")
  (assert (some? (:data gray)) "Should have base64 data")
  (println "   ✓ grayscale test passed!"))

;; Test 5: Draw text
(println "\n5. Testing draw-text:")
(let [with-text (img/draw-text "test.png" {:text "Hello Pod!" :x 10 :y 30 :color "#FF0000"})]
  (println (format "   Text drawn: %dx%d" (:width with-text) (:height with-text)))
  (assert (= 200 (:width with-text)) "Draw-text should preserve dimensions")
  (assert (= 150 (:height with-text)) "Draw-text should preserve dimensions")
  (assert (some? (:data with-text)) "Should have base64 data")
  (println "   ✓ draw-text test passed!"))

(println "\n✅ All new function tests passed!")
