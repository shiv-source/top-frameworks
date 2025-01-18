package main

import (
	"fmt"
	"github.com/shiv-source/frameworkinsights/utils"
	"github.com/shiv-source/markdownTable"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Repository struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	URL         string  `json:"html_url"`
	Description string  `json:"description"`
	Stars       int     `json:"stargazers_count"`
	Forks       int     `json:"forks_count"`
	Watchers    int     `json:"watchers_count"`
	Subscribers int     `json:"subscribers_count"`
	Issues      int     `json:"open_issues_count"`
	Language    string  `json:"language"`
	UpdatedAt   string  `json:"updated_at"`
	Score       float64 `json:"score"`
}

func callApi(url, accessToken string, ch chan<- Repository, wg *sync.WaitGroup) {
	defer wg.Done()
	result, err := utils.MakeAuthenticatedGETRequest[Repository](url, accessToken)
	if err != nil {
		fmt.Printf("Error calling %s: %v\n", url, err)
		return
	}
	ch <- *result
}

func main() {
	startTime := time.Now()
	const txtFileName = "projects.txt"
	const githubUrl = "https://github.com/"
	const githubBaseApiUrl = "https://api.github.com"
	const outputJsonFile = "frameworks.json"
	const templateFile = "template.md"
	const outputTemplateFile = "readme.md"
	accessToken := os.Getenv("MY_GITHUB_ACCESS_TOKEN")

	if accessToken == "" {
		fmt.Fprintln(os.Stderr, "Error: MY_GITHUB_ACCESS_TOKEN is not set")
		os.Exit(1)
	}

	urls := utils.LoadUrlsFromTxtFile(txtFileName)

	ch := make(chan Repository)
	var wg sync.WaitGroup
	repositories := []Repository{}

	for _, url := range urls {
		if strings.HasPrefix(url, githubUrl) {
			repoFullName := strings.TrimPrefix(url, githubUrl)
			repoUrl := fmt.Sprintf("%s/repos/%s", githubBaseApiUrl, repoFullName)
			fmt.Printf("Fetching data from repository => %s\n", repoUrl)
			wg.Add(1)
			go callApi(repoUrl, accessToken, ch, &wg)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	metricRanges := map[string]struct {
		min, max int
	}{
		"Stars":       {min: math.MaxInt, max: math.MinInt},
		"Forks":       {min: math.MaxInt, max: math.MinInt},
		"Watchers":    {min: math.MaxInt, max: math.MinInt},
		"Subscribers": {min: math.MaxInt, max: math.MinInt},
		"Issues":      {min: math.MaxInt, max: math.MinInt},
	}

	for repo := range ch {
		metricRanges["Stars"] = struct {
			min, max int
		}{
			min: int(math.Min(float64(metricRanges["Stars"].min), float64(repo.Stars))),
			max: int(math.Max(float64(metricRanges["Stars"].max), float64(repo.Stars))),
		}

		metricRanges["Forks"] = struct {
			min, max int
		}{
			min: int(math.Min(float64(metricRanges["Forks"].min), float64(repo.Forks))),
			max: int(math.Max(float64(metricRanges["Forks"].max), float64(repo.Forks))),
		}

		metricRanges["Watchers"] = struct {
			min, max int
		}{
			min: int(math.Min(float64(metricRanges["Watchers"].min), float64(repo.Watchers))),
			max: int(math.Max(float64(metricRanges["Watchers"].max), float64(repo.Watchers))),
		}

		metricRanges["Subscribers"] = struct {
			min, max int
		}{
			min: int(math.Min(float64(metricRanges["Subscribers"].min), float64(repo.Subscribers))),
			max: int(math.Max(float64(metricRanges["Subscribers"].max), float64(repo.Subscribers))),
		}

		metricRanges["Issues"] = struct {
			min, max int
		}{
			min: int(math.Min(float64(metricRanges["Issues"].min), float64(repo.Issues))),
			max: int(math.Max(float64(metricRanges["Issues"].max), float64(repo.Issues))),
		}

		repositories = append(repositories, repo)
	}

	for i, repo := range repositories {
		normalizedStars := float64(repo.Stars-metricRanges["Stars"].min) / float64(metricRanges["Stars"].max-metricRanges["Stars"].min)
		normalizedForks := float64(repo.Forks-metricRanges["Forks"].min) / float64(metricRanges["Forks"].max-metricRanges["Forks"].min)
		normalizedWatchers := float64(repo.Watchers-metricRanges["Watchers"].min) / float64(metricRanges["Watchers"].max-metricRanges["Watchers"].min)
		normalizedSubscribers := float64(repo.Subscribers-metricRanges["Subscribers"].min) / float64(metricRanges["Subscribers"].max-metricRanges["Subscribers"].min)
		normalizedIssues := float64(repo.Issues-metricRanges["Issues"].min) / float64(metricRanges["Issues"].max-metricRanges["Issues"].min)
		repositories[i].Score = calculateWeightScore(normalizedStars, normalizedForks, normalizedWatchers, normalizedSubscribers, normalizedIssues)
	}

	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Score > repositories[j].Score
	})

	var tableBody [][]string
	for i, repo := range repositories {
		parsedTime, err := time.Parse(time.RFC3339, repo.UpdatedAt)
		if err != nil {
			panic(err)
		}

		tableBody = append(tableBody, []string{
			fmt.Sprintf("%d", i+1),
			repo.Name,
			fmt.Sprintf("%d", repo.Stars),
			fmt.Sprintf("%d", repo.Forks),
			fmt.Sprintf("%d", repo.Issues),
			repo.Language,
			strings.TrimSpace(repo.Description),
			parsedTime.Format("2006-01-02 15:04:05"),
		})
	}

	previewTableHead := []string{"SL", "Name", "Stars", "Forks", "Issues", "Language"}
	previewTable := markdownTable.CreateMarkdownTable(previewTableHead, tableBody)
	fmt.Println("\n\n" + previewTable + "\n")
	previewTableHead = append(previewTableHead, "Description", "UpdatedAt")
	table := markdownTable.CreateMarkdownTable(previewTableHead, tableBody)

	data := struct {
		Table       string
		LastUpdated string
	}{
		Table:       table,
		LastUpdated: time.Now().Format("January 02, 2006"),
	}
	err := utils.SaveToMarkdown(templateFile, data, outputTemplateFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Markdown generated and saved to => %s\n", outputTemplateFile)
	utils.SaveToJsonFile(repositories, outputJsonFile)
	fmt.Printf("Total repositories fetched => %d\n", len(repositories))
	fmt.Printf("Total execution time: %.3f seconds\n", time.Since(startTime).Seconds())
}

func calculateWeightScore(normalizedStars float64, normalizedForks float64, normalizedWatchers float64, normalizedSubscribers float64, normalizedIssues float64) float64 {
	//weight score settings
	weights := map[string]float64{
		"Stars":       0.4,  // 40% weight
		"Forks":       0.25, // 25% weight
		"Watchers":    0.2,  // 20% weight
		"Subscribers": 0.1,  // 10% weight
		"Issues":      0.05, // 5%  weight
	}

	score := normalizedStars*weights["Stars"] +
		normalizedForks*weights["Forks"] +
		normalizedWatchers*weights["Watchers"] +
		normalizedSubscribers*weights["Subscribers"] +
		normalizedIssues*weights["Issues"]
	return math.Round(score*1000) / 1000.0
}
