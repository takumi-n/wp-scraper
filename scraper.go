package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"net/http"
	"bytes"
	"encoding/json"
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
			articles, err := s.scrapeCategory(url)
			if err != nil {
				errCh <- err
			}

			resultCh <- Result{Category: category, Articles: articles}
		}(url, category)
	}

	resultCount := 0
	for {
		select {
		case result := <-resultCh:
			s.results = append(s.results, result)
			resultCount++

			if resultCount == len(s.config.Categories) {
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

	// create site on demo server
	endpointToCreateSite := s.config.Destination + "/sites/" + s.config.SiteName
	req, err := http.NewRequest(
		"POST",
		endpointToCreateSite,
		bytes.NewBufferString(""),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.config.AuthUsername, s.config.AuthPassword)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	resp.Body.Close()

	// post articles
	endpointToPostArticles := endpointToCreateSite + "/articles/"

	categoryID := 1

	for _, result := range s.results {
		var articleJsons []articleJsonData
		articleID := 1

		for _, article := range result.Articles {
			articleJson := articleJsonData{
				ID:       articleID,
				Title:    article.Title,
				Link:     article.Url,
				Eyecatch: article.Eyecatch,
			}

			articleJsons = append(articleJsons, articleJson)

			articleID++
		}

		categoryJson := categoryJsonData{ID: categoryID, Name: result.Category, Source: "http://example.com"}
		requestJson := requestJsonData{
			Category: categoryJson,
			Articles: articleJsons,
		}

		jsonBytes, err := json.Marshal(requestJson)

		if err != nil {
			return "", err
		}

		req, err := http.NewRequest(
			"POST",
			endpointToPostArticles,
			bytes.NewBuffer(jsonBytes),
		)

		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(s.config.AuthUsername, s.config.AuthPassword)

		client := &http.Client{}
		resp, err := client.Do(req)

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
