package main

import (
	"fmt"
	"github.com/shiv-source/TechTracker/utils"
	"github.com/shiv-source/markdownTable"
	"math"
	"os"
	"path/filepath"
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

type Config struct {
	Id        int    `json:"id"`
	GroupName string `json:"groupName"`
	FilePath  string `json:"filePath"`
}

type Group struct {
	Id             int          `json:"id"`
	GroupName      string       `json:"groupName"`
	InputFilePath  string       `json:"InputFilePath"`
	OutputFilePath string       `json:"OutputFilePath"`
	Repositories   []Repository `json:"repositories"`
}

func main() {
	startTime := time.Now()
	const configFile = "config.json"
	const githubUrl = "https://github.com/"
	const githubBaseApiUrl = "https://api.github.com"
	const outputDir = "data"
	const templateFile = "template.md"
	const outputTemplateFile = "readme.md"
	accessToken := os.Getenv("GITHUB_TOKEN")

	configurations := utils.LoadJSONFromFile[[]Config](configFile)
	if configurations == nil || len(*configurations) == 0 {
		fmt.Println("No configurations found.")
		return
	}

	var (
		configurationWg sync.WaitGroup
		groupsChan      = make(chan Group, len(*configurations))
	)

	for _, config := range *configurations {
		configurationWg.Add(1)
		go func(cfg Config) {
			defer configurationWg.Done()
			urls := utils.LoadUrlsFromTxtFile(cfg.FilePath)

			var (
				repoWg    sync.WaitGroup
				reposChan = make(chan Repository, len(urls))
			)

			for _, url := range urls {
				repoFullName := strings.TrimPrefix(url, githubUrl)
				repoUrl := fmt.Sprintf("%s/repos/%s", githubBaseApiUrl, repoFullName)
				repoWg.Add(1)
				go callApi(repoUrl, accessToken, reposChan, &repoWg)
			}
			repoWg.Wait()
			close(reposChan)

			var repositories []Repository
			for result := range reposChan {
				repositories = append(repositories, result)
			}

			baseFileName := filepath.Base(cfg.FilePath)
			ext := filepath.Ext(baseFileName)
			outputFileName := strings.TrimSuffix(baseFileName, ext)
			outputFilePath := filepath.Join(outputDir, outputFileName+".json")

			groupsChan <- Group{
				Id:             cfg.Id,
				GroupName:      cfg.GroupName,
				Repositories:   getRepositoriesWithScore(repositories),
				InputFilePath:  cfg.FilePath,
				OutputFilePath: outputFilePath,
			}

		}(config)
	}

	configurationWg.Wait()
	close(groupsChan)

	var groups []Group
	for result := range groupsChan {
		groups = append(groups, result)
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Id < groups[j].Id
	})

	saveGroupsToJson(groups, outputDir)
	normalizedAllRepositoryWithScoreAndSave(groups, outputDir)
	createMarkdownTable(groups, templateFile, outputTemplateFile)
	fmt.Printf("Total execution time: %.3f seconds\n", time.Since(startTime).Seconds())
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

func getRepositoriesWithScore(repositories []Repository) []Repository {
	metricRanges := map[string]struct {
		min, max int
	}{
		"Stars":       {min: math.MaxInt, max: math.MinInt},
		"Forks":       {min: math.MaxInt, max: math.MinInt},
		"Watchers":    {min: math.MaxInt, max: math.MinInt},
		"Subscribers": {min: math.MaxInt, max: math.MinInt},
		"Issues":      {min: math.MaxInt, max: math.MinInt},
	}

	for _, repo := range repositories {
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

	return repositories
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
	return math.Round(score*100000) / 100000.0
}

func saveGroupsToJson(groups []Group, outputDir string) {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	for _, group := range groups {
		wg.Add(1)
		go func(group Group) {
			defer wg.Done()
			utils.SaveToJsonFile(group.Repositories, group.OutputFilePath)
		}(group)
	}
	wg.Wait()
	fmt.Println("All files saved.")
}

func normalizedAllRepositoryWithScoreAndSave(groups []Group, outputDir string) {
	var repositories []Repository
	for _, repo := range groups {
		repositories = append(repositories, repo.Repositories...)
	}
	repositories = getRepositoriesWithScore(repositories)
	utils.SaveToJsonFile(repositories, fmt.Sprintf("%s/all.json", outputDir))
}

func createMarkdownTable(groups []Group, templateFile string, outputTemplateFile string) {
	tableHeader := []string{"SL", "Name", "Stars", "Forks", "Issues", "Language", "Description", "UpdatedAt"}
	tableGroup := ""
	for _, group := range groups {
		var tableBody [][]string
		for i, repo := range group.Repositories {
			tableBody = append(tableBody, []string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("[%s](%s)", repo.Name, repo.URL),
				fmt.Sprintf("%d", repo.Stars),
				fmt.Sprintf("%d", repo.Forks),
				fmt.Sprintf("%d", repo.Issues),
				repo.Language,
				strings.TrimSpace(repo.Description),
				getFormattedDateTime(repo.UpdatedAt),
			})
		}
		tableGroup += fmt.Sprintf("## 📋 %s \n\n", group.GroupName)
		tableGroup += fmt.Sprintf("%s \n\n\n", markdownTable.CreateMarkdownTable(tableHeader, tableBody))
	}

	data := struct {
		Table       string
		LastUpdated string
	}{
		Table:       tableGroup,
		LastUpdated: time.Now().Format("January 02, 2006"),
	}
	err := utils.SaveToMarkdown(templateFile, data, outputTemplateFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Markdown generated and saved to => %s\n", outputTemplateFile)

}

func getFormattedDateTime(dateTime string) string {
	var formattedTime string

	if dateTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, dateTime)
		if err != nil {
			panic(err)
		}
		formattedTime = parsedTime.Format("2006-01-02")
	} else {
		formattedTime = ""
	}

	return formattedTime
}
