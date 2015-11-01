package main

import (
	"flag"
	"fmt"
	"gvweb/simplemux"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const g_DataDir = "data/"

var g_supportedTools = [...]string{"dot", "neato", "twopi", "circo", "fdp", "sfdp", "patchwork"}

func homePage(w http.ResponseWriter, r *http.Request, matches []string) {
	http.ServeFile(w, r, "static/index.html")
}

func isSupportedTool(tool string) bool {
	for _, val := range g_supportedTools {
		if tool == val {
			return true
		}
	}
	return false
}

func generateHandler(w http.ResponseWriter, r *http.Request, matches []string) {
	graph := r.FormValue("graphtext")
	imgType := r.FormValue("imagetype")
	tool := r.FormValue("tool")

	if !isSupportedTool(tool) {
		http.Error(w, fmt.Sprintf("Tool '%s' is not supported", tool), http.StatusBadRequest)
		return
	}
	if len(imgType) == 0 {
		http.Error(w, fmt.Sprintf("imagetype is not specified", tool), http.StatusBadRequest)
		return
	}

	if len(graph) == 0 {
		http.Error(w, "Empty input", http.StatusBadRequest)
		return
	}

	result := runGraphviz(tool, graph, imgType)
	if result.err != nil {
		log.Print(result.err)
		http.Error(w, result.err.Error(), http.StatusNotAcceptable)
		return
	} else {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(result.fileName))
		return
	}
}

func accessLogger(l chan string) {
	for {
		log.Print(<-l)
	}
}

func httpLog(l chan string, w http.ResponseWriter, r *http.Request) {

	var remote string
	if len(r.Header["X-Forwarded-For"]) > 0 {
		remote = r.Header["X-Forwarded-For"][0]
	} else {
		remote = r.RemoteAddr
	}
	l <- fmt.Sprintf("%s %s %s", remote, r.Method, r.URL)
}

func httpWrapper(l chan string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Go HTTP handler")
		handler.ServeHTTP(w, r)
		httpLog(l, w, r)
	})
}

func serveHTTP(port string) {
	accessLogChan := make(chan string, 64)
	go accessLogger(accessLogChan)

	reHandler := simplemux.NewRegexpHandler()
	reHandler.AddRoute("^/$", "GET", homePage)
	reHandler.AddRoute("^/generate$", "POST", generateHandler)
	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.Handle("/"+g_DataDir, http.FileServer(http.Dir(".")))
	http.Handle("/", reHandler)

	log.Printf("gvweb(%s) listening at port %s\n", g_Version, port)
	err := http.ListenAndServe(":"+port, httpWrapper(accessLogChan, http.DefaultServeMux))
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

var g_Port = flag.Int("port", 12345, "port number to listen on")
var g_CleanupInterval = flag.Int("purge", 24*60*60, "Remove saved graphs that are older than this amount in seconds. 0 to keep them forever.")
var g_Version = "DEVELOPMENT"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n%s v.%s\n", os.Args[0], g_Version)
	}

	flag.Parse()

	runtime.GOMAXPROCS(3)

	if *g_CleanupInterval > 0 {
		interval := time.Duration(*g_CleanupInterval) * time.Second
		initPurge(g_DataDir, interval)
		log.Printf("Purging data older than %v\n", interval)
	}

	serveHTTP(fmt.Sprintf("%d", *g_Port))
}
