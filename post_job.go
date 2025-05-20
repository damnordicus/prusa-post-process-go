package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FilamentPayload struct {
	Filename     string `json:"filename"`
	FilamentUsed string `json:"filament_used"`
	PrinterModel string `json:"printer_model"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <raw_file_path>")
		return
	}

	rawFile := os.Args[1]
	logFile := "C:\\Users\\PhoenixSpark\\Documents\\post_job.log" // Adjust for Windows path later

	logOutput, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer logOutput.Close()
	log.SetOutput(logOutput)
	log.Println("Reading Temp file")

	// Strip .pp extension
	gcodeFile := strings.TrimSuffix(rawFile, ".pp")
	outputFilename := filepath.Base(os.Getenv("SLIC3R_PP_OUTPUT_NAME")) // Might be empty in Windows
	log.Println("Output file:", outputFilename)

	// Open raw file
	file, err := os.Open(gcodeFile)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var filamentUsed string
	var printerModel string
	newUsed := "should never happen"
	scanner := bufio.NewScanner(file)

	// Regex for matching "filament.*used" and extracting numeric values
	//re := regexp.MustCompile("[0-9.]+")
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)
		if strings.Contains(lower, "filament") && strings.Contains(lower, "used") {
			if strings.Contains(lower, "filament used [g]") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					filamentUsed = strings.TrimSpace(parts[1])
				}
			} 
		}
		if strings.Contains(lower, "printer") && strings.Contains(lower, "model") {
			if strings.Contains(lower, "; printer_model") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					printerModel = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(lower, "printer_model=") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					printerModel = strings.TrimSpace(parts[1])
				}
			}
		}		
		if filamentUsed != "" && printerModel != "" {
			break
		}
	}

	if len(filamentUsed) == 0 {
		log.Println("Filament used extraction failed")
		return
	}

	log.Println("Filament used:", filamentUsed)
	log.Println("Printer model:", printerModel)
	log.Println("New used:", newUsed)

	// Build JSON
	payload := FilamentPayload{
		Filename:     outputFilename,
		FilamentUsed: filamentUsed,
		PrinterModel: printerModel,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error creating JSON:", err)
		return
	}
	log.Println(string(jsonData))

	// POST request
	resp, err := http.Post("https://filmanager.apps.travisspark.com/api/pendingJob", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending POST:", err)
		return
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
}