package scraper

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type (
	Class struct {
		Title    ClassSelector `yaml:"title"`
		URL      ClassSelector `yaml:"url"`
		Eyecatch ClassSelector `yaml:"eyecatch"`
	}

	ClassSelector struct {
		CSS           string `yaml:"css"`
		Target        string `yaml:"target"`
		AdditionalCSS string `yaml:"additional_css"`
		Regex         string `yaml:"regex"`
	}

	Config struct {
		Destination     string            `yaml:"destination"`
		SiteName        string            `yaml:"site_name"`
		AuthUsername    string            `yaml:"auth_username"`
		AuthPassword    string            `yaml:"auth_password"`
		BaseURL         string            `yaml:"base_url"`
		Categories      map[string]string `yaml:"categories"`
		ArticleSelector string            `yaml:"article_selector"`
		Class           Class             `yaml:"classes"`
	}
)

// Make Config struct from config file
func ReadConfig(filePath string) (*Config, error) {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return marshalYAMLByte(buf)
}

func marshalYAMLByte(buffer []byte) (*Config, error) {
	var c Config
	err := yaml.Unmarshal(buffer, &c)
	if err != nil {
		return nil, err
	}

	return &c, err
}

func normalizeConfig(config *Config) {
	if config.BaseURL[len(config.BaseURL)-1:] == "/" {
		config.BaseURL = config.BaseURL[:len(config.BaseURL)-1]
	}

	if config.Destination[len(config.Destination)-1:] == "/" {
		config.Destination = config.Destination[:len(config.Destination)-1]
	}
}
