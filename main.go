package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var client *http.Client

type Repo struct {
	Name           string `json:"name"`
	Stars          int    `json:"stargazers_count"`
	Forks          int    `json:"forks_count"`
	Issues         int    `json:"open_issues_count"`
	URL            string `json:"html_url"`
	Description    string `json:"description"`
	Private        bool   `json:"private"`
	DefaultBranch  string `json:"default_branch"`
	CommitsUrl     string `json:"commits_url"`
	Language       string `json:"language"`
	LastCommitDate string
}

type HeadCommit struct {
	Sha    string `json:"sha"`
	Commit struct {
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

func main() {
	godotenv.Load()
	fileName := "project-list.txt"
	urls := loadProjects(fileName)
	githubUrl := "https://github.com/"
	github_access_token := os.Getenv("MY_GITHUB_ACCESS_TOKEN")

	ch := make(chan Repo)
	var wg sync.WaitGroup

	projects := []Repo{}

	fmt.Println("Fetching data from Github")
	for _, url := range urls {

		if strings.HasPrefix(url, githubUrl) {

			repoFullName := strings.TrimPrefix(url, githubUrl)

			repoUrl := fmt.Sprintf("https://api.github.com/repos/%s", repoFullName)

			commitUrl := fmt.Sprintf("https://api.github.com/repos/%s/commits", repoFullName)

			wg.Add(1)
			go getRepoInfo(repoUrl, github_access_token, commitUrl, ch, &wg)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for res := range ch {
		projects = append(projects, res)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Stars > projects[j].Stars
	})

	saveToReadme(projects, "readme.md")

	fmt.Println("Total Projects: =", len(urls))
	fmt.Println("Done!!!")
}

func loadProjects(fileName string) []string {

	byteCodes, err := ioutil.ReadFile(fileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	projects := strings.Split(string(byteCodes), "\n")

	return projects
}

func getData(url string, token string, target interface{}) error {

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	if err != nil {
		fmt.Println(err)
	}

	client = &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func getCommitInfo(commitUrl string, token string) string {

	var commit []HeadCommit

	err := getData(commitUrl, token, &commit)

	if err != nil {
		fmt.Println(err)
	}

	return commit[0].Commit.Committer.Date.Format("2006-01-02 15:04:05")
}

func getRepoInfo(repoUrl string, github_access_token string, commitUrl string, ch chan<- Repo, wg *sync.WaitGroup) {
	defer wg.Done()

	var repo Repo

	err := getData(repoUrl, github_access_token, &repo)

	if err != nil {
		fmt.Println(err)
	}

	repo.LastCommitDate = getCommitInfo(commitUrl, github_access_token)

	ch <- repo
}

func saveToReadme(repos []Repo, fileName string) {

	header := `# Top Frameworks
## A list of top frameworks ranked by stars on github.  
Please update the project-list.txt file.

| SL| Name  | Stars| Forks| Issues | Language | Description | Last Commit |
| --| ------| -----| ---- | ------ | -------- | ----------- | ----------- |
`
	footer := "\n### Last updated at : %s\n"
	readme, err := os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)

	if err != nil {
		fmt.Println(err)
	}
	defer readme.Close()

	readme.WriteString(header)

	for i, repo := range repos {

		description := strings.ReplaceAll(repo.Description, "\n", "")

		readme.WriteString(fmt.Sprintf("| %d | [%s](%s) | %d | %d | %d | %s | %s | %s |\n", i+1, repo.Name, repo.URL, repo.Stars, repo.Forks, repo.Issues, repo.Language, description, repo.LastCommitDate))
	}
	readme.WriteString(fmt.Sprintf(footer, time.Now().Format("2006-01-02 15:04:05")))
	fmt.Println("Successfully written the Markdown file")
}
