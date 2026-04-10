package main

import (
	frontenddist "github.com/EugeneNail/vox/ui/internal"
	"io/fs"
	"log"
	"net/http"
	"os"
)

func main() {
	distFileSystem, err := fs.Sub(frontenddist.FileSystem, "dist")
	if err != nil {
		log.Fatalf("resolving embedded frontend dist: %v", err)
	}

	indexHtml, err := fs.ReadFile(distFileSystem, "index.html")
	if err != nil {
		log.Fatalf("reading embedded index.html: %v", err)
	}

	fileServer := http.FileServerFS(distFileSystem)
	webServer := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			response.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = response.Write(indexHtml)
			return
		}

		if _, err := fs.Stat(distFileSystem, request.URL.Path[1:]); err == nil {
			fileServer.ServeHTTP(response, request)
			return
		}

		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = response.Write(indexHtml)
	})

	address := ":" + uiPort()
	log.Printf("ui service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}

func uiPort() string {
	port := os.Getenv("UI_PORT")
	if port == "" {
		return "5173"
	}

	return port
}
