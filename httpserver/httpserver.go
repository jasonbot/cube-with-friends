package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"text/template"
)

//go:embed static/**
var content embed.FS

var frontpagetemplate *template.Template

func init() {
	f, err := content.Open("static/index.template")
	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	frontpagetemplate, err = template.New("indexpage").Parse(string(b))
	if err != nil {
		panic(err)
	}
}

type RenderFlags struct {
	ShowCanvas       bool
	ConnectionString string
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		hostname := r.URL.Hostname()
		if hostname == "" {
			hostname = "bad"
		}
		username := r.Form.Get("username")
		connectionstring := []string{username, "", hostname, "25565"}
		c, _ := json.Marshal(connectionstring)

		frontpagetemplate.Execute(w, RenderFlags{ShowCanvas: true, ConnectionString: string(c)})
		return
	}

	w.WriteHeader(200)
	w.Header().Add("content-type", "text/html")
	frontpagetemplate.Execute(w, RenderFlags{ShowCanvas: false})
}

func ServeHttp(cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	http.Handle("/static/", http.FileServer(http.FS(content)))
	http.HandleFunc("/", mainPage)

	listenOn := "0.0.0.0:5555"

	var err error
	go func() {
		log.Println("Listening to HTTP requests on", listenOn)
		err = http.ListenAndServe(listenOn, nil)
		if err != nil {
			cancel()
		}
	}()

	<-c.Done()
	log.Println("Finishing up HTTP server...")
	return err
}
