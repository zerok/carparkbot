package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	token           string
	channelName     string
	mappingFilePath string
	port            string
)

var plateChars *regexp.Regexp = regexp.MustCompile(`[^a-zA-Z0-9]`)

func init() {
	flag.StringVar(&token, "token", "", "Command token for validation")
	flag.StringVar(&channelName, "channel", "", "Supported channel")
	flag.StringVar(&mappingFilePath, "mapping", "", "Path to a mapping file")
	flag.StringVar(&port, "port", "8080", "Port to run on")
	flag.Parse()

	if token == "" {
		log.Fatal("Please specify a token using -token")
	}
	if channelName == "" {
		log.Fatal("Please specify a channel using -channel")
	}
	if mappingFilePath == "" {
		log.Fatal("Please specify a mapping file using -mapping")
	}
}

func main() {
	store, err := newMappingStore(mappingFilePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = store.Start(); err != nil {
		log.Fatal(err.Error())
	}
	defer store.Stop()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Print(err.Error())
			http.Error(w, "Failed to parse request", 500)
		}
		if !isInChannel(r, channelName) {
			sendInlineResponse(w, fmt.Sprintf("This command is only supported in #%s", channelName))
			return
		}
		plate := extractPlate(r)
		if plate == "" {
			sendInlineResponse(w, "Please specify a plate number.")
			return
		}
		holder, found := store.LookupPlate(plate)
		if !found {
			sendInlineResponse(w, fmt.Sprintf("Holder for plate %s not found", plate))
			return
		}
		user := r.Form["user_name"][0]
		if user == holder {
			sendInlineResponse(w, "You have blocked yourself. Congratulations!")
			return
		}
		sendChannelResponse(w, fmt.Sprintf("@%s: Your :car: is blocking @%s! Please move it.", holder, user))
	})
	log.Printf("Starting server on port %s", port)
	if err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err.Error())
	}
}

func extractPlate(r *http.Request) string {
	texts, ok := r.Form["text"]
	if !ok {
		return ""
	}
	if len(texts) != 1 {
		return ""
	}
	return plateChars.ReplaceAllString(strings.TrimSpace(texts[0]), "")
}

func isInChannel(r *http.Request, requiredChannel string) bool {
	names, ok := r.Form["channel_name"]
	if !ok {
		return false
	}
	if len(names) < 1 {
		return false
	}
	return names[0] == requiredChannel
}

func sendInlineResponse(w http.ResponseWriter, msg string) {
	w.Write([]byte(msg))
}

type channelResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func sendChannelResponse(w http.ResponseWriter, msg string) {
	resp := channelResponse{ResponseType: "in_channel", Text: msg}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to send channel response: %s", err.Error())
	}
}
