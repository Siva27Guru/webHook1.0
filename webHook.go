package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type IncomingData struct {
	Ev     string `json:"ev"`
	Et     string `json:"et"`
	Id     string `json:"id"`
	Uid    string `json:"uid"`
	Mid    string `json:"mid"`
	T      string `json:"t"`
	P      string `json:"p"`
	L      string `json:"l"`
	Sc     string `json:"sc"`
	Atrk1  string `json:"atrk1"`
	Atrv1  string `json:"atrv1"`
	Atrt1  string `json:"atrt1"`
	Atrk2  string `json:"atrk2"`
	Atrv2  string `json:"atrv2"`
	Atrt2  string `json:"atrt2"`
	Uatrk1 string `json:"uatrk1"`
	Uatrv1 string `json:"uatrv1"`
	Uatrt1 string `json:"uatrt1"`
	Uatrk2 string `json:"uatrk2"`
	Uatrv2 string `json:"uatrv2"`
	Uatrt2 string `json:"uatrt2"`
	Uatrk3 string `json:"uatrk3"`
	Uatrv3 string `json:"uatrv3"`
	Uatrt3 string `json:"uatrt3"`
}

type TransformedData struct {
	Event           string               `json:"event"`
	EventType       string               `json:"event_type"`
	AppID           string               `json:"app_id"`
	UserID          string               `json:"user_id"`
	MessageID       string               `json:"message_id"`
	PageTitle       string               `json:"page_title"`
	PageURL         string               `json:"page_url"`
	BrowserLanguage string               `json:"browser_language"`
	ScreenSize      string               `json:"screen_size"`
	Attributes      map[string]Attribute `json:"attributes"`
	Traits          map[string]Trait     `json:"traits"`
}

type Attribute struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Trait struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

var dataChannel = make(chan IncomingData)

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	go worker()
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var incoming IncomingData
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dataChannel <- incoming
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func worker() {
	for incoming := range dataChannel {
		go func(incoming IncomingData) {
			transformed := transformData(incoming)
			sendToWebhook(transformed)
		}(incoming)
	}
}

func transformData(incoming IncomingData) TransformedData {
	attributes := map[string]Attribute{
		incoming.Atrk1: {Value: incoming.Atrv1, Type: incoming.Atrt1},
		incoming.Atrk2: {Value: incoming.Atrv2, Type: incoming.Atrt2},
	}
	traits := map[string]Trait{
		incoming.Uatrk1: {Value: incoming.Uatrv1, Type: incoming.Uatrt1},
		incoming.Uatrk2: {Value: incoming.Uatrv2, Type: incoming.Uatrt2},
		incoming.Uatrk3: {Value: incoming.Uatrv3, Type: incoming.Uatrt3},
	}
	return TransformedData{
		Event:           incoming.Ev,
		EventType:       incoming.Et,
		AppID:           incoming.Id,
		UserID:          incoming.Uid,
		MessageID:       incoming.Mid,
		PageTitle:       incoming.T,
		PageURL:         incoming.P,
		BrowserLanguage: incoming.L,
		ScreenSize:      incoming.Sc,
		Attributes:      attributes,
		Traits:          traits,
	}
}

func sendToWebhook(data TransformedData) {
	url := "https://webhook.site/638f632b-d301-4e1a-b90a-27079f1b5976" // Replace with your webhook URL
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK response: %d", resp.StatusCode)
	}
}
