package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GitHub API からのレスポンスをマッピングする構造体
type GithubResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct {
							ContributionCount int    `json:"contributionCount"`
							Date              string `json:"date"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
}

// GitHub API への認証トークン
var TOKEN string

// GitHub のユーザー名
var USER = "Fuuma0000"

// GitHub API のエンドポイント URL
var URL = "https://api.github.com/graphql"

// GraphQL のクエリ
var QUERY = fmt.Sprintf(`
{
  user(login: "%s") {
    contributionsCollection {
      contributionCalendar {
        weeks {
          contributionDays {
            contributionCount
            date
          }
        }
      }
    }
  }
}
`, USER)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	TOKEN = os.Getenv("GITHUB_API_TOKEN")

	// GitHub API へのリクエストを作成する
	requestBody, err := json.Marshal(map[string]string{"query": QUERY})
	if err != nil {
		log.Fatal(err)
	}

	request, err := http.NewRequest("POST", URL, strings.NewReader(string(requestBody)))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", TOKEN))

	// GitHub API へのリクエストを実行する
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// レスポンスを GithubResponse 構造体にマッピングする
	var githubReponse GithubResponse
	if err := json.NewDecoder(response.Body).Decode(&githubReponse); err != nil {
		log.Fatal(err)
	}

	// 今日のコントリビューション数を取得し、草を生やしていない場合にメッセージを出力する
	if !isTodayContributed(githubReponse) {
		fmt.Println("今日の草を生やしていません。")
		f, err := os.OpenFile("failureDay.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		today := githubReponse.Data.User.ContributionsCollection.ContributionCalendar.Weeks[0].ContributionDays[0].Date
		fmt.Fprintln(f, today)

		return
	}
}

// 今日のコントリビューションがあるかどうかを判定する関数
func isTodayContributed(response GithubResponse) bool {
	todayContribution := response.Data.User.ContributionsCollection.ContributionCalendar.Weeks[0].ContributionDays[0]
	return todayContribution.ContributionCount > 0
}
