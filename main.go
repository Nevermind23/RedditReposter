package main

import (
	"encoding/json"
	"fmt"
	"github.com/mmcdole/gofeed/rss"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64
}

func main() {
	credentials := getCredentials()
	feed := getRssFeed()
	i := 0
	for {
		now := time.Now()
		token := credentials.AccessToken
		if now.Unix() > credentials.ExpiresAt {
			token = redditRefreshToken(credentials.RefreshToken)
		}
		link, name := getLatestNews(feed, i)
		submitReddit(token, link, name)
		fmt.Println("Submitted in", time.Since(now))
		time.Sleep(10 * time.Second)
		if i == len(feed.Items) {
			feed = getRssFeed()
			i = 0
		} else {
			i++
		}
	}
}

func submitReddit(token, link, name string) {
	submitUrl := "https://oauth.reddit.com/api/submit"
	client := http.Client{
		Timeout: time.Second * 5,
	}

	form := url.Values{}
	form.Add("sr", "SharingNews")
	form.Add("kind", "link")
	form.Add("title", name)
	form.Add("url", link+"?utm_source=reddit")
	form.Add("resubmit", "true")
	form.Add("api_type", "json")

	req, err := http.NewRequest("POST", submitUrl, strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Auto Sharing News (by /u/ImmaChallenger)")
	req.Header.Add("Authorization", "Bearer "+token)
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(req.Body)
}

func getRssFeed() *rss.Feed {
	feedUrl := "https://stardiapost.com/feed?take=all"
	client := &http.Client{}
	req, err := http.NewRequest("GET", feedUrl, nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fp := rss.Parser{}
	rssFeed, _ := fp.Parse(strings.NewReader(string(body)))

	return rssFeed
}

func getLatestNews(rssFeed *rss.Feed, i int) (string, string) {
	rand.Seed(time.Now().UTC().UnixNano())
	item := rssFeed.Items[i]
	return item.Link, item.Title
}

func redditRefreshToken(token string) string {
	link := "https://www.reddit.com/api/v1/access_token"
	client := http.Client{
		Timeout: time.Second * 5,
	}

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", token)

	req, err := http.NewRequest("POST", link, strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth("gr-KnD1-PYRoupHBIMD7kg", "7x8ffQNFucxwnCbgdckND57ekw7rxQ")
	req.Header.Set("User-Agent", "Auto Sharing News (by /u/ImmaChallenger)")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(req.Body)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	result := Credentials{}

	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	result.ExpiresAt = time.Now().Unix() + 3400
	byteRes, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("./data/credentials.json", byteRes, 0644)
	if err != nil {
		panic(err)
	}
	return result.AccessToken
}

func getCredentials() Credentials {
	jsonFile, err := os.Open("./data/credentials.json")
	if err != nil {
		panic(err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	result := Credentials{}
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		panic(err)
	}

	return result
}
