package main

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
	"strings"

	transit "github.com/babashka/transit-go"
	"github.com/minikomi/pod-golang-image/babashka"
	podimage "github.com/minikomi/pod-golang-image/image"
)

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read message: %v\n", err)
			return
		}

		switch message.Op {
		case "describe":
			handleDescribe(message)
		case "invoke":
			handleInvoke(message)
		default:
			err := fmt.Errorf("unknown op: %s", message.Op)
			babashka.WriteErrorResponse(message, err)
		}
	}
}

func handleDescribe(message *babashka.Message) {
	describeResponse := &babashka.DescribeResponse{
		Format: "transit+json",
		Namespaces: []babashka.Namespace{
			{
				Name: "co.poyo.pod-golang-image",
				Vars: []babashka.Var{
					{Name: "info"},
					{Name: "resize"},
					{Name: "to-base64"},
					{Name: "rotate"},
					{Name: "flip"},
					{Name: "crop"},
					{Name: "grayscale"},
					{Name: "draw-text"},
				},
			},
		},
	}
	babashka.WriteDescribeResponse(describeResponse)
}

func handleInvoke(message *babashka.Message) {
	switch message.Var {
	case "co.poyo.pod-golang-image/info":
		handleInfo(message)
	case "co.poyo.pod-golang-image/resize":
		handleResize(message)
	case "co.poyo.pod-golang-image/to-base64":
		handleToBase64(message)
	case "co.poyo.pod-golang-image/rotate":
		handleRotate(message)
	case "co.poyo.pod-golang-image/flip":
		handleFlip(message)
	case "co.poyo.pod-golang-image/crop":
		handleCrop(message)
	case "co.poyo.pod-golang-image/grayscale":
		handleGrayscale(message)
	case "co.poyo.pod-golang-image/draw-text":
		handleDrawText(message)
	default:
		err := fmt.Errorf("unknown var: %s", message.Var)
		babashka.WriteErrorResponse(message, err)
	}
}

func parseArgs(args string) (*list.List, error) {
	reader := strings.NewReader(args)
	decoder := transit.NewDecoder(reader)
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return value.(*list.List), nil
}

func respond(message *babashka.Message, response interface{}) {
	buf := bytes.NewBufferString("")
	encoder := transit.NewEncoder(buf, false)
	if err := encoder.Encode(response); err != nil {
		babashka.WriteErrorResponse(message, err)
	} else {
		babashka.WriteInvokeResponse(message, buf.String())
	}
}

func handleInfo(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() != 1 {
		babashka.WriteErrorResponse(message, fmt.Errorf("info expects 1 argument, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	result, err := podimage.Info(path)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleToBase64(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() != 1 {
		babashka.WriteErrorResponse(message, fmt.Errorf("to-base64 expects 1 argument, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	result, err := podimage.ToBase64(path)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleResize(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() < 1 || argsList.Len() > 2 {
		babashka.WriteErrorResponse(message, fmt.Errorf("resize expects 1 or 2 arguments, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	// Parse options if provided
	var opts podimage.ResizeOptions
	if argsList.Len() == 2 {
		optsElement := argsList.Front().Next()
		if optsMap, ok := optsElement.Value.(map[interface{}]interface{}); ok {
			opts = parseResizeOptions(optsMap)
		} else {
			babashka.WriteErrorResponse(message, fmt.Errorf("second argument must be a map (options)"))
			return
		}
	}

	result, err := podimage.Resize(path, opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func parseResizeOptions(opts map[interface{}]interface{}) podimage.ResizeOptions {
	result := podimage.ResizeOptions{
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

func handleRotate(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() < 1 || argsList.Len() > 2 {
		babashka.WriteErrorResponse(message, fmt.Errorf("rotate expects 1 or 2 arguments, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	// Parse options if provided
	opts := make(map[interface{}]interface{})
	if argsList.Len() == 2 {
		optsElement := argsList.Front().Next()
		if optsMap, ok := optsElement.Value.(map[interface{}]interface{}); ok {
			opts = optsMap
		} else {
			babashka.WriteErrorResponse(message, fmt.Errorf("second argument must be a map (options)"))
			return
		}
	}

	result, err := podimage.Rotate(path, opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleFlip(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() < 1 || argsList.Len() > 2 {
		babashka.WriteErrorResponse(message, fmt.Errorf("flip expects 1 or 2 arguments, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	// Parse options if provided
	opts := make(map[interface{}]interface{})
	if argsList.Len() == 2 {
		optsElement := argsList.Front().Next()
		if optsMap, ok := optsElement.Value.(map[interface{}]interface{}); ok {
			opts = optsMap
		} else {
			babashka.WriteErrorResponse(message, fmt.Errorf("second argument must be a map (options)"))
			return
		}
	}

	result, err := podimage.Flip(path, opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleCrop(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() != 2 {
		babashka.WriteErrorResponse(message, fmt.Errorf("crop expects 2 arguments, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	optsElement := argsList.Front().Next()
	var opts map[interface{}]interface{}
	if optsMap, ok := optsElement.Value.(map[interface{}]interface{}); ok {
		opts = optsMap
	} else {
		babashka.WriteErrorResponse(message, fmt.Errorf("second argument must be a map (options)"))
		return
	}

	result, err := podimage.Crop(path, opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleGrayscale(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() != 1 {
		babashka.WriteErrorResponse(message, fmt.Errorf("grayscale expects 1 argument, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	result, err := podimage.Grayscale(path)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}

func handleDrawText(message *babashka.Message) {
	argsList, err := parseArgs(message.Args)
	if err != nil {
		babashka.WriteErrorResponse(message, fmt.Errorf("failed to parse args: %w", err))
		return
	}

	if argsList.Len() != 2 {
		babashka.WriteErrorResponse(message, fmt.Errorf("draw-text expects 2 arguments, got %d", argsList.Len()))
		return
	}

	path, ok := argsList.Front().Value.(string)
	if !ok {
		babashka.WriteErrorResponse(message, fmt.Errorf("first argument must be a string (path)"))
		return
	}

	optsElement := argsList.Front().Next()
	var opts map[interface{}]interface{}
	if optsMap, ok := optsElement.Value.(map[interface{}]interface{}); ok {
		opts = optsMap
	} else {
		babashka.WriteErrorResponse(message, fmt.Errorf("second argument must be a map (options)"))
		return
	}

	result, err := podimage.DrawText(path, opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
		return
	}

	respond(message, result)
}
