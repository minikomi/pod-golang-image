#!/usr/bin/env bb

;; Simple test to verify pod works
(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[pod.poyo.image :as img])

(println "Pod loaded successfully!")
(println "\nTesting pod.poyo.image functions:\n")

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

(println "\n✅ All tests passed!")
