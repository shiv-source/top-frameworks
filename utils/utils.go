package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func LoadUrlsFromTxtFile(txtFileName string) []string {
	dataBytes, err := os.ReadFile(txtFileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lines := strings.Split(string(dataBytes), "\n")
	re := regexp.MustCompile(`https?://[^\s]+`)
	uniqueUrls := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if url := re.FindString(line); url != "" {
			uniqueUrls[url] = true
		}
	}

	var urls []string
	for url := range uniqueUrls {
		urls = append(urls, url)
	}

	return urls
}

func MakeAuthenticatedGETRequest[T any](url, accessToken string) (*T, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error un-marshalling response body: %v", err)
	}

	return &result, nil
}

func SaveToJsonFile[T any](data T, fileName string) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data to JSON: %w", err)
	}

	err = os.WriteFile(fileName, jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON to file: %w", err)
	}

	fmt.Printf("Json data has been saved into => %s\n", fileName)

	return nil
}

func SaveToMarkdown(templatePath string, data interface{}, outputPath string) error {
	mdTemplate, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	parsedTemplate, err := template.New("markdown").Parse(string(mdTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var output bytes.Buffer
	if err := parsedTemplate.Execute(&output, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(outputPath, output.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func LoadJSONFromFile[T any](fileName string) *T {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("failed to read file: %v\n", err)
		os.Exit(1)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Printf("failed to unmarshal JSON: %v\n", err)
		os.Exit(1)
	}

	return &result
}
