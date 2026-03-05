#!/usr/bin/env bb

;; Example usage of pod-golang-image
(require '[babashka.pods :as pods]
         '[clojure.string :as str])

(pods/load-pod "./pod-golang-image")
(require '[co.poyo.pod-golang-image :as img])

(println "=== pod-golang-image Example ===")
(println)

;; Example 1: Get image info
(println "1. Get image info (fast, header-only):")
(let [info (img/info "test.png")]
  (println "   Width:" (:width info))
  (println "   Height:" (:height info))
  (println "   Type:" (:media-type info)))
(println)

;; Example 2: Resize with max-edge constraint
(println "2. Resize with max-edge=100:")
(let [result (img/resize "test.png" {:max-edge 100})]
  (println "   New dimensions:" (:width result) "x" (:height result))
  (println "   Output type:" (:media-type result))
  (println "   Base64 data:" (str (subs (:data result) 0 50) "...")))
(println)

;; Example 3: Resize with max-width
(println "3. Resize with max-width=80:")
(let [result (img/resize "test.png" {:max-width 80})]
  (println "   New dimensions:" (:width result) "x" (:height result))
  (println "   Aspect ratio preserved:"
           (= (/ 200 150) (/ (:width result) (:height result)))))
(println)

;; Example 4: Resize with format conversion
(println "4. Resize and convert to JPEG:")
(let [result (img/resize "test.png" {:max-edge 120 :format "jpeg" :quality 95})]
  (println "   New dimensions:" (:width result) "x" (:height result))
  (println "   Output type:" (:media-type result))
  (println "   Quality: 95"))
(println)

;; Example 5: Encode to base64 without resizing
(println "5. Encode to base64 without resizing:")
(let [result (img/to-base64 "test.png")]
  (println "   Dimensions:" (:width result) "x" (:height result))
  (println "   Base64 length:" (count (:data result)) "characters"))
(println)

(println "✅ All examples completed successfully!")
