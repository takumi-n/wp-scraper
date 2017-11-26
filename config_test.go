package scraper

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"fmt"
)

func TestConfigMarshal(t *testing.T) {
	ast := assert.New(t)
	yaml := `
destination: https://example.com/destination
site_name: Google
auth_username: Joe
auth_password: secret
base_url: https://example.com/base
categories:
     Food: food_url
     Medical: medical_url
     Lifestyle: lifecycle_url
     Fashion: fashion_url
article_selector: article
classes:
    title:
        css: title-css
        target: text
    url:
        css: url-css
        target: attribute
        additional_css: href
        regex: abc(.+)
    eyecatch:
        css: eyecatch-css
        target: attribute
        additional_css: src
`

	c, err := marshalYAMLByte([]byte(yaml))
	ast.Nil(err)

	ast.Equal("Joe", c.AuthUsername)
	ast.Equal("food_url", c.Categories["Food"])
	ast.Equal("title-css", c.Class.Title.CSS)
	ast.Equal("", c.Class.Title.AdditionalCSS)
	ast.Equal("href", c.Class.URL.AdditionalCSS)
	fmt.Println("REGEX = " + c.Class.URL.Regex)
	ast.Equal("abc(.+)", c.Class.URL.Regex)
}

func TestNormalizeConfig(t *testing.T) {
	ast := assert.New(t)

	c := new(Config)
	c.Destination = "http://example.com/destination/"
	c.BaseURL = "http://example.com/base/"

	normalizeConfig(c)

	ast.Equal("http://example.com/destination", c.Destination)
	ast.Equal("http://example.com/base", c.BaseURL)
}
