package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FilamentPayload struct {
	Filename     string    `json:"filename"`
	FilamentUsed []float64 `json:"filament_used"`
	PrinterModel string    `json:"printer_model"`
}

func strToFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		log.Printf("Filament used extraction failed, %v\n", err)
		return math.NaN(), err
	}

	f = math.Round(f*100) / 100

	return f, nil
}

func trimQuotes(s string) string {
	return strings.TrimSpace(strings.Trim(s, `"`))
}

func splitPrefix(s string) (string, error) {
	parts := strings.Split(strings.TrimSpace(s), "=")

	if len(parts) != 2 {
		return "", errors.New("invalid prefix")
	}

	return strings.TrimSpace(parts[1]), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <raw_file_path>")
		return
	}

	rawFile := os.Args[1]
	//logFile := "C:\\Users\\PhoenixSpark\\Documents\\post_job.log" // Adjust for Windows path later
	logDir, _ := os.Getwd()
	logFile := logDir + "/logs"
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

	filamentUsed := make([]float64, 0, 5)
	var printerModel string

	scanner := bufio.NewScanner(file)

	// Regex for matching "filament.*used" and extracting numeric values
	//re := regexp.MustCompile("[0-9.]+")
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)

		// figure out bgcode vs gcode
		if lower[:1] == ";" {
			// if gcode remove ;
			lower = lower[2:]
		}
		//fmt.Println(lower)
		if strings.HasPrefix(lower, "filament used [g]") {
			// check prefix for `filament used [g]`
			weightsStr, err := splitPrefix(lower)
			if err != nil {
				log.Println("Error parsing filament used [g]:", err)
				return
			}

			if strings.Contains(weightsStr, ",") {
				weights := strings.Split(trimQuotes(weightsStr), ",")
				for _, weight := range weights {
					f, err := strToFloat(weight)
					if err != nil {
						log.Printf("Filament used extraction failed, %v\n", err)
						return
					}
					filamentUsed = append(filamentUsed, f)
				}
			} else {
				weight := trimQuotes(weightsStr)

				f, err := strToFloat(weight)
				if err != nil {
					log.Printf("Filament used extraction failed, %v\n", err)
					return
				}
				filamentUsed = append(filamentUsed, f)
			}
		}

		if strings.HasPrefix(lower, "printer_model") {
			printerModel, err = splitPrefix(lower)
			if err != nil {
				log.Printf("Filament used extraction failed, %v\n", err)
			}
			printerModel = strings.ToUpper(printerModel)
		}

		if len(filamentUsed) > 0 && printerModel != "" {
			break
		}
	}

	if len(filamentUsed) == 0 {
		log.Println("Filament used extraction failed")
		return
	}
	log.Println("Filament used:", filamentUsed)
	log.Println("Printer model:", printerModel)
	//fmt.Printf("Filament used: %v\n", filamentUsed)
	//fmt.Printf("Printer model: %v\n", printerModel)

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
