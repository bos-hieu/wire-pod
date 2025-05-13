package main

import (
	"encoding/json"
	"fmt"
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
	"set preference",
	"get preference",
	"delete preference",
}

var Name = "Personal Information"

type PersonalInfo struct {
	Name        string            `json:"name"`
	Preferences map[string]string `json:"preferences"`
}

var personalData PersonalInfo

// Constants for file paths and messages
const (
	dataDirName  = "data"
	dataFileName = "personal.json"
)

// Error messages
const (
	errNoNameSet     = "I don't know your name yet. You can tell me your name by saying 'set my name' followed by your name."
	errNoPreferences = "You haven't set any preferences yet."
	errInvalidInput  = "I didn't catch that. Could you please repeat it?"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("[DEBUG] Initializing personal plugin")
	
	personalData = PersonalInfo{
		Name:        "",
		Preferences: make(map[string]string),
	}
	
	if err := loadPersonalData(); err != nil {
		log.Printf("[ERROR] Failed to initialize personal data: %v", err)
	}
}

func getDataFilePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))
	dataDir := filepath.Join(baseDir, "chipper", "plugins", "personal", dataDirName)
	
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %v", err)
	}
	
	return filepath.Join(dataDir, dataFileName), nil
}

func loadPersonalData() error {
	log.Println("[DEBUG] Loading personal data")
	
	dataFile, err := getDataFilePath()
	if err != nil {
		return fmt.Errorf("failed to get data file path: %v", err)
	}

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		log.Println("[DEBUG] Personal data file does not exist, creating new file")
		return savePersonalData()
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("failed to read personal data file: %v", err)
	}

	if err := json.Unmarshal(data, &personalData); err != nil {
		return fmt.Errorf("failed to unmarshal personal data: %v", err)
	}
	
	log.Printf("[DEBUG] Successfully loaded personal data: %+v", personalData)
	return nil
}

func savePersonalData() error {
	log.Println("[DEBUG] Saving personal data")
	
	dataFile, err := getDataFilePath()
	if err != nil {
		return fmt.Errorf("failed to get data file path: %v", err)
	}
	
	data, err := json.MarshalIndent(personalData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal personal data: %v", err)
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write personal data file: %v", err)
	}
	
	log.Printf("[DEBUG] Successfully saved personal data: %+v", personalData)
	return nil
}

func stripOutTriggerWords(s string) string {
	log.Printf("[DEBUG] Stripping trigger words from: %s", s)
	
	triggerWords := []string{"simon says", "repeat", "hey vector", "ok vector"}
	result := strings.ToLower(s)
	
	for _, word := range triggerWords {
		result = strings.Replace(result, word, "", 1)
	}
	
	result = strings.TrimSpace(result)
	log.Printf("[DEBUG] Stripped result: %s", result)
	return result
}

func handleNameQuery() (string, string) {
	if personalData.Name == "" {
		return "intent_imperative_praise", errNoNameSet
	}
	return "intent_imperative_praise", fmt.Sprintf("Your name is %s", personalData.Name)
}

func handleNameUpdate(name string) (string, string) {
	if name == "" {
		return "intent_imperative_praise", errInvalidInput
	}
	
	personalData.Name = name
	if err := savePersonalData(); err != nil {
		log.Printf("[ERROR] Failed to save name update: %v", err)
		return "intent_imperative_praise", "I had trouble saving your name. Please try again."
	}
	
	return "intent_imperative_praise", fmt.Sprintf("I'll remember that your name is %s", name)
}

func handlePreferencesQuery() (string, string) {
	if len(personalData.Preferences) == 0 {
		return "intent_imperative_praise", errNoPreferences
	}
	
	var response strings.Builder
	response.WriteString("Here are your preferences: ")
	
	for key, value := range personalData.Preferences {
		response.WriteString(fmt.Sprintf("%s is set to %s, ", key, value))
	}
	
	return "intent_imperative_praise", strings.TrimSuffix(response.String(), ", ")
}

func handlePreferenceUpdate(text string) (string, string) {
	parts := strings.Split(text, "set preference")
	if len(parts) != 2 {
		return "intent_imperative_praise", "Please specify a preference in the format 'set preference [key] to [value]'"
	}
	
	prefText := strings.TrimSpace(parts[1])
	prefParts := strings.Split(prefText, " to ")
	
	if len(prefParts) != 2 {
		return "intent_imperative_praise", "Please specify a preference in the format 'set preference [key] to [value]'"
	}
	
	key := strings.TrimSpace(prefParts[0])
	value := strings.TrimSpace(prefParts[1])
	
	if key == "" || value == "" {
		return "intent_imperative_praise", errInvalidInput
	}
	
	personalData.Preferences[key] = value
	if err := savePersonalData(); err != nil {
		log.Printf("[ERROR] Failed to save preference update: %v", err)
		return "intent_imperative_praise", "I had trouble saving your preference. Please try again."
	}
	
	return "intent_imperative_praise", fmt.Sprintf("I've set %s to %s", key, value)
}

func Action(transcribedText string, botSerial string, guid string, target string) (string, string) {
	log.Printf("[DEBUG] Action called with text: %s, botSerial: %s, guid: %s, target: %s", 
		transcribedText, botSerial, guid, target)
	
	text := stripOutTriggerWords(transcribedText)
	
	switch {
	case strings.Contains(text, "what's my name") || strings.Contains(text, "who am i"):
		return handleNameQuery()
		
	case strings.Contains(text, "set my name") || strings.Contains(text, "remember my name"):
		parts := strings.Split(text, "set my name")
		if len(parts) > 1 {
			return handleNameUpdate(strings.TrimSpace(parts[1]))
		}
		return "intent_imperative_praise", errInvalidInput
		
	case strings.Contains(text, "what are my preferences"):
		return handlePreferencesQuery()
		
	case strings.Contains(text, "set preference"):
		return handlePreferenceUpdate(text)
		
	default:
		return "intent_imperative_praise", "I'm not sure how to help with that. You can ask me about your name or preferences."
	}
} 
