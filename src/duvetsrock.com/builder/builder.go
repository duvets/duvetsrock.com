package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "log"
import "os"
import "path"
import "time"
import "html/template"
import "gopkg.in/yaml.v1"

type BandData struct {
	Gigs      []Gig
	Locations map[string]Location
	Songs     []Song
}
type Gig struct {
	StartTime time.Time
	Duration  string
	Location  string
}
type Location struct {
	Title   string
	Address string
	City    string
	State   string
	Zip     string
}
type Song struct {
	Title  string
	Artist string
	Year   int
}

func decodeYAML(path string, result interface{}) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contents, result)
	if err != nil {
		return fmt.Errorf("Error parsing YAML file %q: %v", path, err)
	}
	fmt.Printf("%v", result)
	return nil
}

func decodeJSON(path string, result interface{}) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contents, result)
	if err != nil {
		return fmt.Errorf("Error parsing JSON file %q: %v", path, err)
	}
	fmt.Printf("%v", result)
	return nil
}

func buildWebsite() error {
	root := os.Getenv("GOPATH")
	var err error

	//	s := []Song{}
	//	err = decodeYAML("/tmp/foo.txt", &s)
	//	return err

	//s := BandData{}
	//err = decodeJSON("/tmp/foo.txt", &s)
	//return err

	bandData := BandData{}
	err = decodeYAML(path.Join(root, "data/data.yaml"), &bandData)
	if err != nil {
		return err
	}
	return nil
	t := template.Must(template.ParseFiles(
		path.Join(root, "web/index.html")))
	err = t.Execute(os.Stdout, bandData)
	if err != nil {
		log.Fatalf("Bad template failed: %s", err)
	}
	return nil
}

func main() {
	err := buildWebsite()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
