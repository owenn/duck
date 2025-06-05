package main

import (
	// "context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	//	"strings"
	"time"
)

//go:embed assets/*.png
var duckImages embed.FS

//go:embed html/*.html
var htmlFS embed.FS

//go:embed assets/duckIndex.json
var duckIndex string

func Hello() string {
	return "Hello, world"
}

type Ducks struct {
	ID    int    `json:"id"`
	NAME  string `json:"name"`
	FNAME string `json:"fname"`
	URL   string `json:"URL"`
	DESCR string `json:"description"`
}

var DuckDB []Ducks

// AppState holds shared application data
type AppState struct {
	AppName  string
	Version  string
	ImageDir string
}

func listAllFiles() {
	files, err := fs.ReadDir(duckImages, "assets")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Embedded files:")
	for _, file := range files {
		fmt.Println(file.Name())
	}

}

func loadDucksIntoDB() {
	// Convert JSON array to slice of structs
	err := json.Unmarshal([]byte(duckIndex), &DuckDB)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for i := 0; i < len(DuckDB); i++ {

		DuckDB[i].ID = i
		DuckDB[i].URL = "image/" + regexp.MustCompile(`[^/\\]+$`).FindString(DuckDB[i].FNAME)

		duck := DuckDB[i]
		log.Println("fname: ", duck.FNAME)
		log.Println("description: ", duck.DESCR)
		log.Println("ID: ", duck.ID)
		log.Println("NAME: ", duck.NAME)
		log.Println("URL: ", duck.URL)
	}

}

func loadHTMLIntoMemory() {
	// Convert JSON array to slice of structs
	err := json.Unmarshal([]byte(duckIndex), &DuckDB)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

}

// ListFiles retrieves the list of file names from a given directory
func listFiles(directory string) ([]string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	return fileList, nil
}

// ListFilesHandler handles HTTP requests and returns file names as JSON
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	directory := "./assets" // Change this to your target directory

	fileList, err := listFiles(directory)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileList)
}

func getDucks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(DuckDB)

	return

}

func getDuckImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")

	duckURL := chi.URLParam(r, "duckImageName")
	log.Println("DuckURL = ", duckURL)

	duckGraphic, err := duckImages.ReadFile("assets/" + duckURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)

	}

	w.Write([]byte(duckGraphic))
	//http.ServeFile(w, r, "assets/"+duckURL)
}

func listFromEmbedded(ff embed.FS, startDir string) {
	entries, err := fs.ReadDir(ff, startDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	for _, entry := range entries {
		fmt.Println("Name:", entry.Name(), "IsDir:", entry.IsDir())
	}
}

func getHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	page := chi.URLParam(r, "pageid")
	//	page = strings.TrimSuffix(page, "\n")
	log.Println("page = [" + page + "]")
	//listFromEmbedded(htmlFS, "html")

	htmlPage, err := htmlFS.ReadFile("html/" + page)
	if err != nil {
		fmt.Println("Error reading file:", err)
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}
	log.Println("serve file" + string(htmlPage))

	w.Write([]byte(htmlPage))
	//http.ServeFile(w, r, "html/")

}

func getParticularDuck(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "duckid")
	fmt.Print("the id = ", id, "\n")

	return
}

func getDuckIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(duckIndex))
}

// DelayMiddleware introduces a delay based on the "delay" URL parameter
func DelayMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxDelay := 30000 // 30 seconds max
		// Get the "delay" parameter from the URL
		delayStr := r.URL.Query().Get("delay")
		log.Println("Got a delay of " + delayStr)

		// Convert delay to an integer (milliseconds)
		if delayMs, err := strconv.Atoi(delayStr); err == nil && delayMs > 0 {

			if delayMs > maxDelay {
				delayMs = maxDelay
			}

			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			log.Println("Got a delay of " + delayStr + "Now ended")
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create a new router
	r := chi.NewRouter()

	loadDucksIntoDB()
	// A good base middleware stack - will neeed to work out what
	// it offers
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(6 * time.Second))

	// Apply the delay middleware globally (for all routes)
	r.Use(DelayMiddleware)

	// Define a simple handler for the root path
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	r.Route("/ducks", func(r chi.Router) {
		r.Get("/", getDucks)                  // GET /articles
		r.Get("/{duckid}", getParticularDuck) // GET /articles/01-16-2017
		r.Get("/index", getDuckIndex)
		r.Get("/image/{duckImageName}", getDuckImage)
		r.Get("/html/{pageid}", getHTML)
	})

	// Start the server
	http.ListenAndServe(":2001", r)
}
