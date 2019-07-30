package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-yaml/yaml"
)

type Pubspec struct {
	Dependencies    map[string]interface{} `yaml:"dependencies,omitempty"`
	DevDependencies map[string]interface{} `yaml:"dev_dependencies,omitempty"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("you must specify a path to one pubspec.yaml")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	pubspec := Pubspec{}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(bytes, &pubspec)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("-- dependencies updates --\n")
	iterateDependencies(pubspec.Dependencies)

	fmt.Println("-- dev_dependencies updates --\n")
	iterateDependencies(pubspec.DevDependencies)
}

func iterateDependencies(deps map[string]interface{}) {
	for pkg, version := range deps {
		version, ok := version.(string)
		if !ok {
			continue
		}

		version = strings.TrimPrefix(version, "^")

		newestVersion, err := getNewestVersion(pkg)
		if err != nil {
			log.Println(err)
		}

		if version != newestVersion {
			fmt.Printf("%s: %s\n\tnew version: %s\n\tchangelog: %s#-changelog-tab-\n\n",
				pkg, version, newestVersion, getPackageUrl(pkg))
		}
	}
}

func getNewestVersion(pkg string) (version string, err error) {
	// Request the HTML page.
	res, err := http.Get(getPackageUrl(pkg))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	title := doc.Find(".package-header").Find(".title").Text()
	version = strings.Split(title, " ")[1]

	return version, nil
}

func getPackageUrl(pkg string) string {
	return "https://pub.dev/packages/" + pkg
}
