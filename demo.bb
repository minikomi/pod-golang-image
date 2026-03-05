#!/usr/bin/env bb

;; Demo: Process an image and display all its metadata

(require '[babashka.pods :as pods])
(pods/load-pod "./pod-golang-image")
(require '[co.poyo.pod-golang-image :as img])

(def image-path "test.png")

(println "=================================")
(println "  Image Processing Demo")
(println "=================================\n")

;; Step 1: Get original image info
(println "📷 Original Image:")
(let [info (img/info image-path)]
  (println (format "   Size: %dx%d pixels" (:width info) (:height info)))
  (println (format "   Type: %s" (:media-type info))))

(println)

;; Step 2: Create thumbnail
(println "🔽 Creating thumbnail (max 64px):")
(let [thumb (img/resize image-path {:max-edge 64 :format "jpeg" :quality 85})]
  (println (format "   New size: %dx%d pixels" (:width thumb) (:height thumb)))
  (println (format "   Format: %s" (:media-type thumb)))
  (println (format "   Data size: %d bytes" (/ (count (:data thumb)) 4 3)))) ; rough base64->bytes

(println)

;; Step 3: Multiple constraint test
(println "📐 Testing constraint combinations:")
(doseq [[desc opts] [["Max edge 100px" {:max-edge 100}]
                      ["Max width 80px" {:max-width 80}]
                      ["Max height 50px" {:max-height 50}]]]
  (let [result (img/resize image-path opts)]
    (println (format "   %s → %dx%d"
                     desc (:width result) (:height result)))))

(println)
(println "✅ All operations completed successfully!")
