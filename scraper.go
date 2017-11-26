package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"net/http"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type (
	Scraper struct {
		config    Config
		isVerbose bool
		results   []Result
	}

	Result struct {
		Category string
		Articles []Article
	}

	Article struct {
		Title    string
		Url      string
		Eyecatch string
	}
)

// Create new scraper with config and verbose mode
func NewScraper(config Config, isVerbose bool) *Scraper {
	s := new(Scraper)
	s.config = config
	s.isVerbose = isVerbose
	return s
}

// Start scraping with limit
func (s *Scraper) Scrape(limit int) error {
	resultCh := make(chan Result)
	errCh := make(chan error)

	for categoryPath, category := range s.config.Categories {
		url := s.config.BaseURL + categoryPath
		go func(url, category string) {
			s.printVerbosely("[" + category + "] Scraping " + url + " ...")
			articles, err := s.scrapeCategory(url)
			if err != nil {
				errCh <- err
			}

			s.printVerbosely("[" + category + "] Result: len = " + strconv.Itoa(len(articles)))

			resultCh <- Result{Category: category, Articles: articles}
		}(url, category)
	}

	for {
		select {
		case result := <-resultCh:
			s.results = append(s.results, result)

			if len(s.results) == len(s.config.Categories) {
				if s.isVerbose {
					count := 0
					for _, result := range s.results {
						count += len(result.Articles)
					}
					s.printVerbosely(fmt.Sprintf("Total result: category len = %v, article len = %v", len(s.results), count))
				}
				return nil
			}

		case err := <-errCh:
			return err
		}
	}
}

// Send scraped data to destination server
// Return created url
func (s *Scraper) SendToServer() (string, error) {
	// create site on demo server
	endpointToCreateSite := s.config.Destination + "/sites/" + s.config.SiteName
	req, err := createHttpRequest("POST", endpointToCreateSite, []byte(""))

	if err != nil {
		return "", err
	}

	s.setupRequest(req)

	resp, err := (&http.Client{}).Do(req)

	if err != nil {
		return "", err
	}
	resp.Body.Close()

	// post articles
	type (
		categoryJsonData struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Source string `json:"source"`
		}

		articleJsonData struct {
			ID       int    `json:"id"`
			Title    string `json:"title"`
			Link     string `json:"link"`
			Eyecatch string `json:"eyecatch"`
		}

		requestJsonData struct {
			Category categoryJsonData  `json:"category"`
			Articles []articleJsonData `json:"articles"`
		}
	)

	endpointToPostArticles := endpointToCreateSite + "/articles/"

	articleID := 1
	categoryID := 1

	for _, result := range s.results {
		categoryJson := categoryJsonData{ID: categoryID, Name: result.Category, Source: "http://example.com"}
		requestJson := requestJsonData{
			Category: categoryJson,
		}

		for _, article := range result.Articles {
			articleJson := articleJsonData{
				ID:       articleID,
				Title:    article.Title,
				Link:     article.Url,
				Eyecatch: article.Eyecatch,
			}

			requestJson.Articles = append(requestJson.Articles, articleJson)
			articleID++
		}

		s.printVerbosely("[" + result.Category + "] POST to demo server")

		jsonBytes, err := json.Marshal(requestJson)

		if err != nil {
			return "", err
		}

		req, err := createHttpRequest("POST", endpointToPostArticles, jsonBytes)

		if err != nil {
			return "", err
		}

		s.setupRequest(req)

		resp, err := (&http.Client{}).Do(req)

		if err != nil {
			return "", err
		}
		resp.Body.Close()

		categoryID++
	}

	return endpointToPostArticles, nil
}

func (s *Scraper) scrapeCategory(url string) ([]Article, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var articles []Article
	doc.Find(s.config.ArticleSelector).Each(func(_ int, selection *goquery.Selection) {
		if err != nil {
			return
		}

		var title, url, eyecatch string

		switch s.config.Class.Title.Target {
		case "text":
			title = selection.Find(s.config.Class.Title.CSS).Text()
		case "attribute":
			t, ok := selection.Find(s.config.Class.Title.CSS).Attr(s.config.Class.Title.AdditionalCSS)
			if ok {
				title = t
			}
		}

		if pattern := s.config.Class.Title.Regex; pattern != "" {
			r := regexp.MustCompile(pattern)
			group := r.FindSubmatch([]byte(title))

			title = string(group[1])
		}

		switch s.config.Class.URL.Target {
		case "text":
			url = selection.Find(s.config.Class.URL.CSS).Text()
		case "attribute":
			u, ok := selection.Find(s.config.Class.URL.CSS).Attr(s.config.Class.URL.AdditionalCSS)
			if ok {
				url = u
			}
		}

		if pattern := s.config.Class.URL.Regex; pattern != "" {
			r := regexp.MustCompile(pattern)
			group := r.FindSubmatch([]byte(url))

			url = string(group[1])
		}

		switch s.config.Class.Eyecatch.Target {
		case "text":
			eyecatch = selection.Find(s.config.Class.Eyecatch.CSS).Text()
		case "attribute":
			e, ok := selection.Find(s.config.Class.Eyecatch.CSS).Attr(s.config.Class.Eyecatch.AdditionalCSS)
			if ok {
				eyecatch = e
			}
		}

		if pattern := s.config.Class.Eyecatch.Regex; pattern != "" {
			r := regexp.MustCompile(pattern)
			group := r.FindSubmatch([]byte(eyecatch))

			eyecatch = string(group[1])
		}

		article := Article{Title: title, Url: url, Eyecatch: eyecatch}
		articles = append(articles, article)
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (s *Scraper) printVerbosely(message string) {
	if s.isVerbose {
		fmt.Println(message)
	}
}

func (s *Scraper) setupRequest(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.AuthUsername, s.config.AuthPassword)
}

func createHttpRequest(method, url string, body []byte) (*http.Request, error) {
	return http.NewRequest(
		method,
		url,
		bytes.NewBuffer(body),
	)
}
