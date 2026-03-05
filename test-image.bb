#!/usr/bin/env bb

;; Comprehensive test suite for pod-golang-image
(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[pod.poyo.image :as img])

(println "Pod loaded successfully!")
(println "\nTesting pod.poyo.image functions (8 total):\n")

;; Test 1: info
(println "1. Testing info function:")
(let [info-result (img/info "test.png")]
  (println "   Result:" info-result)
  (assert (= 200 (:width info-result)) "Width should be 200")
  (assert (= 150 (:height info-result)) "Height should be 150")
  (assert (= "image/png" (:media-type info-result)) "Media type should be image/png")
  (println "   ✓ info test passed!"))

;; Test 2: resize with max-edge
(println "\n2. Testing resize with max-edge:")
(let [resize-result (img/resize "test.png" {:max-edge 100})]
  (println "   Result dimensions:" (:width resize-result) "x" (:height resize-result))
  (assert (<= (:width resize-result) 100) "Width should be <= 100")
  (assert (<= (:height resize-result) 100) "Height should be <= 100")
  (assert (some? (:data resize-result)) "Should have base64 data")
  (println "   ✓ resize test passed!"))

;; Test 3: to-base64
(println "\n3. Testing to-base64:")
(let [b64-result (img/to-base64 "test.png")]
  (println "   Result dimensions:" (:width b64-result) "x" (:height b64-result))
  (assert (= 200 (:width b64-result)) "Width should be 200")
  (assert (= 150 (:height b64-result)) "Height should be 150")
  (assert (some? (:data b64-result)) "Should have base64 data")
  (let [data-len (count (:data b64-result))]
    (println "   Base64 data length:" data-len "characters")
    (assert (> data-len 100) "Base64 data should be substantial"))
  (println "   ✓ to-base64 test passed!"))

;; Test 4: rotate
(println "\n4. Testing rotate:")
(let [result (img/rotate "test.png" {:degrees 90})]
  (println "   Result dimensions:" (:width result) "x" (:height result))
  (assert (= 150 (:width result)) "90° should swap width/height")
  (assert (= 200 (:height result)) "90° should swap width/height")
  (println "   ✓ rotate test passed!"))

;; Test 5: flip
(println "\n5. Testing flip:")
(let [result (img/flip "test.png" {:direction "horizontal"})]
  (println "   Result dimensions:" (:width result) "x" (:height result))
  (assert (= 200 (:width result)) "Flip should preserve dimensions")
  (assert (= 150 (:height result)) "Flip should preserve dimensions")
  (println "   ✓ flip test passed!"))

;; Test 6: crop
(println "\n6. Testing crop:")
(let [result (img/crop "test.png" {:x 10 :y 10 :width 80 :height 60})]
  (println "   Result dimensions:" (:width result) "x" (:height result))
  (assert (= 80 (:width result)) "Crop should match requested width")
  (assert (= 60 (:height result)) "Crop should match requested height")
  (println "   ✓ crop test passed!"))

;; Test 7: grayscale
(println "\n7. Testing grayscale:")
(let [result (img/grayscale "test.png")]
  (println "   Result dimensions:" (:width result) "x" (:height result))
  (assert (= 200 (:width result)) "Grayscale should preserve dimensions")
  (assert (= 150 (:height result)) "Grayscale should preserve dimensions")
  (println "   ✓ grayscale test passed!"))

;; Test 8: draw-text
(println "\n8. Testing draw-text:")
(let [result (img/draw-text "test.png" {:text "Test" :x 10 :y 20 :color "#FF0000"})]
  (println "   Result dimensions:" (:width result) "x" (:height result))
  (assert (= 200 (:width result)) "Draw-text should preserve dimensions")
  (assert (= 150 (:height result)) "Draw-text should preserve dimensions")
  (println "   ✓ draw-text test passed!"))

(println "\n✅ All 8 tests passed!")
