package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "os"
import "path"
import "regexp"
import "strings"
import "time"
import "html/template"
import "github.com/go-yaml/yaml"

type BandData struct {
	Gigs      []Gig
	Locations map[string]Location
	Songs     []Song
}
type Gig struct {
	StartTime time.Time
	Duration  time.Duration
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

func get(mapAny interface{}, key string) string {
	m := mapAny.(map[interface{}]interface{})
	result := m[key]
	if result == nil {
		return ""
	}
	return result.(string)

}

var timeRegex = regexp.MustCompile(`(\d\d\d\d-\d\d-\d\d \d\d:\d\d:(?:\d\d)?) ([\w/]+)`)

func parseTime(str string) (time.Time, error) {
	m := timeRegex.FindStringSubmatch(str)
	if m != nil {
		location, err := time.LoadLocation(m[2])
		if err != nil {
			return time.Time{}, err
		}
		return time.ParseInLocation("2006-01-02 15:04:05", m[1], location)
	}
	return time.Time{}, fmt.Errorf("Unable to parse time %q", str)
}

func (g *Gig) SetYAML(tag string, value interface{}) bool {
	startTime, err := parseTime(get(value, "startTime"))
	if err != nil {
		//fmt.Println(err)
	}
	duration, err := time.ParseDuration(get(value, "duration"))
	if err != nil {
		//fmt.Printf("Error setting duration %q %v\n", get(value, "duration"), err)
	}
	g.StartTime = startTime
	g.Duration = duration
	g.Location = get(value, "location")
	return true
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
	return nil
}

var funcMap = template.FuncMap{
	"formatTime": func(formatString string, t time.Time) string {
		return t.Format(formatString)
	},
}

type config struct {
	root string
}

var errNotExist = fmt.Errorf("File does not exist")

func writePageContent(config config, filePath string, w http.ResponseWriter) error {
	fullPath := path.Join(config.root, filePath, "web")
	if stat, err := os.Stat(fullPath); os.IsNotExist(err) || stat.IsDir() {
		fullPath = path.Join(fullPath, "index.html")
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errNotExist
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	bandData := BandData{}
	err := decodeYAML(path.Join(config.root, "data/data.yaml"), &bandData)
	if err != nil {
		return err
	}

	templateStr, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	t, err := template.New(filePath).Funcs(funcMap).Parse(string(templateStr))
	if err != nil {
		return err
	}
	err = t.Execute(w, bandData)
	if err != nil {
		return err
	}
	return nil
}

func serveWebsite(config config) error {
	fileServer := http.FileServer(http.Dir(path.Join(config.root, "web")))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := writePageContent(config, req.URL.Path, w)
		if err == errNotExist {
			fileServer.ServeHTTP(w, req)
			return
		}
    if err != nil {
      w.Header().Add("Content-Type", "text/plain")
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(fmt.Sprintf("%v", err)))
    }
	})
  port := 8080
	s := &http.Server{
		Addr:           fmt.Sprintf(":%v", port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
  hostname, err := os.Hostname()
  if err != nil {
    hostname = "localhost"
  }
  fmt.Printf("Started server at http://%v:%v\n", strings.ToLower(hostname), port)

	return s.ListenAndServe()
}

func main() {
	config := config{
		root: os.Getenv("GOPATH"),
	}

	err := serveWebsite(config)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
