package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	token           string
	channelName     string
	enableDM        bool
	apiToken        string
	mappingFilePath string
	port            string
)

var plateChars *regexp.Regexp = regexp.MustCompile(`[^a-zA-Z0-9]`)

func init() {
	flag.StringVar(&token, "token", "", "Command token for validation")
	flag.StringVar(&apiToken, "api-token", "", "Token required for sending direct messages")
	flag.BoolVar(&enableDM, "dm", false, "Also DM the car holder")
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
	if enableDM && apiToken == "" {
		log.Fatal("You have to specify an API token using -api-token if you want to send direct messages.")
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
	if mappingFilePath == "" {
		http.HandleFunc("/mapping/", func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			if r.Method != "POST" {
				http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
				return
			}
			if err := store.UpdateStore(r.Body); err != nil {
				log.Print(err.Error())
				if err == errInvalidCSV {
					http.Error(w, "Invalid CSV", http.StatusBadRequest)
					return
				}
				http.Error(w, "Server error", http.StatusInternalServerError)
			}
		})
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			r.Body.Close()
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
			return
		}
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
		// if user == holder {
		// 	sendInlineResponse(w, "You have blocked yourself. Congratulations!")
		// 	return
		// }
		sendChannelResponse(w, fmt.Sprintf("<@%s>: Your :car: is blocking <@%s>! Please move it.", holder, user))
		if enableDM && apiToken != "" {
			msg := fmt.Sprintf("Your :car: is blocking <@%s>! Please move it.", user)
			sendDM(apiToken, holder, msg)
		}
	})
	log.Printf("Starting server on port %s", port)
	if err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err.Error())
	}
}

func sendDM(token, blockingPerson, msg string) error {
	resp, err := http.PostForm("https://slack.com/api/chat.postMessage", url.Values{
		"token":   {token},
		"channel": {fmt.Sprintf("@%s", blockingPerson)},
		"text":    {msg},
	})
	if err != nil {
		return err
	}
	defer ioutil.NopCloser(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("Unexpected response code received during DM-sending: %d", resp.StatusCode)
	}
	return nil
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
