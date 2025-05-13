package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var Utterances = []string{
	"what's my name",
	"who am i",
	"what are my preferences",
	"set my name",
	"update my preferences",
	"remember my name",
}

var Name = "Personal Information"

type PersonalInfo struct {
	Name       string            `json:"name"`
	Preferences map[string]string `json:"preferences"`
}

var personalData PersonalInfo

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("[DEBUG] Initializing personal plugin")
	
	// Initialize personal data
	personalData = PersonalInfo{
		Name: "",
		Preferences: make(map[string]string),
	}
	loadPersonalData()
}

func loadPersonalData() {
	log.Println("[DEBUG] Loading personal data")
	
	dataDir := filepath.Join("chipper", "plugins", "personal", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("[ERROR] Failed to create data directory: %v", err)
		return
	}
	
	dataFile := filepath.Join(dataDir, "personal.json")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		log.Println("[DEBUG] Personal data file does not exist, creating new file")
		savePersonalData()
		return
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		log.Printf("[ERROR] Failed to read personal data file: %v", err)
		return
	}

	if err := json.Unmarshal(data, &personalData); err != nil {
		log.Printf("[ERROR] Failed to unmarshal personal data: %v", err)
		return
	}
	
	log.Printf("[DEBUG] Successfully loaded personal data: %+v", personalData)
}

func savePersonalData() {
	log.Println("[DEBUG] Saving personal data")
	
	dataDir := filepath.Join("chipper", "plugins", "personal", "data")
	dataFile := filepath.Join(dataDir, "personal.json")
	
	data, err := json.MarshalIndent(personalData, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal personal data: %v", err)
		return
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		log.Printf("[ERROR] Failed to write personal data file: %v", err)
		return
	}
	
	log.Printf("[DEBUG] Successfully saved personal data: %+v", personalData)
}

func stripOutTriggerWords(s string) string {
	log.Printf("[DEBUG] Stripping trigger words from: %s", s)
	
	result := strings.Replace(s, "simon says", "", 1)
	result = strings.Replace(result, "repeat", "", 1)
	result = strings.TrimSpace(result)
	
	log.Printf("[DEBUG] Stripped result: %s", result)
	return result
}

func Action(transcribedText string, botSerial string, guid string, target string) (string, string) {
	log.Printf("[DEBUG] Action called with text: %s, botSerial: %s, guid: %s, target: %s", 
		transcribedText, botSerial, guid, target)
	
	text := strings.ToLower(stripOutTriggerWords(transcribedText))
	
	switch {
	case strings.Contains(text, "what's my name") || strings.Contains(text, "who am i"):
		log.Println("[DEBUG] Processing name query")
		if personalData.Name == "" {
			log.Println("[DEBUG] No name found in personal data")
			return "intent_imperative_praise", "I don't know your name yet. You can tell me your name by saying 'set my name' followed by your name."
		}
		log.Printf("[DEBUG] Returning name: %s", personalData.Name)
		return "intent_imperative_praise", "Your name is " + personalData.Name

	case strings.Contains(text, "set my name") || strings.Contains(text, "remember my name"):
		log.Println("[DEBUG] Processing name update request")
		parts := strings.Split(text, "set my name")
		if len(parts) > 1 {
			name := strings.TrimSpace(parts[1])
			if name != "" {
				log.Printf("[DEBUG] Setting new name: %s", name)
				personalData.Name = name
				savePersonalData()
				return "intent_imperative_praise", "I'll remember that your name is " + name
			}
		}
		log.Println("[DEBUG] Failed to extract name from request")
		return "intent_imperative_praise", "I didn't catch your name. Could you please repeat it?"

	case strings.Contains(text, "what are my preferences"):
		log.Println("[DEBUG] Processing preferences query")
		if len(personalData.Preferences) == 0 {
			log.Println("[DEBUG] No preferences found")
			return "intent_imperative_praise", "You haven't set any preferences yet."
		}
		response := "Here are your preferences: "
		for key, value := range personalData.Preferences {
			response += key + " is set to " + value + ", "
		}
		log.Printf("[DEBUG] Returning preferences: %s", response[:len(response)-2])
		return "intent_imperative_praise", response[:len(response)-2]

	case strings.Contains(text, "update my preferences"):
		log.Println("[DEBUG] Processing preferences update request")
		// This is a placeholder for preference updates
		// You can extend this to handle specific preference updates
		return "intent_imperative_praise", "I'm still learning how to update preferences. Please try again later."

	default:
		log.Printf("[DEBUG] No matching intent found for text: %s", text)
		return "intent_imperative_praise", "I'm not sure how to help with that. You can ask me about your name or preferences."
	}
} 
