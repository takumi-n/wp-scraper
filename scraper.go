package scraper

import "github.com/PuerkitoBio/goquery"

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
		Content  string
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
func (s *Scraper) Scrape(limit int) ([]Result, error) {
	resultCh := make(chan Result)
	errCh := make(chan error)

	for _, category := range s.config.Categories {
		url := s.config.BaseURL + category
		go func() {
			articles, err := s.scrapeCategory(url)
			if err != nil {
				errCh <- err
			}

			resultCh <- Result{Category: category, Articles: articles}
		}()
	}

	var resultCount int
	for {
		select {
		case result := <-resultCh:
			s.results = append(s.results, result)
			resultCount++

			if resultCount == len(s.results) {
				return s.results, nil
			}

		case err := <-errCh:
			return []Result{}, err
		}
	}
}

// Send scraped data to destination server
func (s *Scraper) SendToServer() error {
	return nil
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

		var title, url, eyecatch, content string

		switch s.config.Class.Title.Target {
		case "text":
			title = selection.Find(s.config.Class.Title.CSS).Text()
		case "attribute":
			t, ok := selection.Find(s.config.Class.Title.CSS).Attr(s.config.Class.Title.AdditionalCSS)
			if ok {
				title = t
			}
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

		switch s.config.Class.Eyecatch.Target {
		case "text":
			eyecatch = selection.Find(s.config.Class.Eyecatch.CSS).Text()
		case "attribute":
			e, ok := selection.Find(s.config.Class.Eyecatch.CSS).Attr(s.config.Class.Eyecatch.AdditionalCSS)
			if ok {
				eyecatch = e
			}
		}

		var contentDoc *goquery.Document
		contentDoc, err = goquery.NewDocument(url)

		switch s.config.Class.Content.Target {
		case "text":
			content = contentDoc.Find(s.config.Class.Content.CSS).Text()
		case "attribute":
			c, ok := contentDoc.Find(s.config.Class.Content.CSS).Attr(s.config.Class.Content.AdditionalCSS)
			if ok {
				content = c
			}
		}

		article := Article{Title: title, Url: url, Eyecatch: eyecatch, Content: content}
		articles = append(articles, article)
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}
