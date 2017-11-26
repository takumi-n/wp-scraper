package scraper

import "github.com/PuerkitoBio/goquery"

type (
	Scraper struct {
		config    Config
		isVerbose bool
		results   []result
	}

	result struct {
		category string
		articles []article
	}

	article struct {
		title    string
		url      string
		eyecatch string
		content  string
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
	resultCh := make(chan result)
	errCh := make(chan error)

	for _, category := range s.config.Categories {
		url := s.config.BaseURL + category
		go func() {
			articles, err := s.scrapeCategory(url)
			if err != nil {
				errCh <- err
			}

			resultCh <- result{category: category, articles: articles}
		}()
	}

	var resultCount int
	for {
		select {
		case result := <-resultCh:
			s.results = append(s.results, result)
			resultCount++

			if resultCount == len(s.results) {
				return nil
			}

		case err := <-errCh:
			return err
		}
	}
}

// Send scraped data to destination server
func (s *Scraper) SendToServer() {

}

func (s *Scraper) scrapeCategory(url string) ([]article, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var articles []article
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

		article := article{title: title, url: url, eyecatch: eyecatch, content: content}
		articles = append(articles, article)
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}
