package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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
			hostname = "localhost"
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

func connectionBanner(port int) error {
	interfaces, err := net.Interfaces()

	if err != nil {
		return err
	}

	log.Println("===========================================")
	log.Println("One of these URLs should work for connecting:")
	for _, i := range interfaces {
		addrs, err := i.Addrs()

		if err != nil {
			return err
		}

		for _, addr := range addrs {
			ipAddress := addr.String()
			ap := strings.Split(ipAddress, "/")
			ipaddr := ap[0]
			if strings.Contains(ipaddr, ":") {
				ipaddr = fmt.Sprintf("[%s]", ipaddr)
			}

			if ipaddr == "127.0.0.1" || ipaddr == "[::1]" {
				continue
			}

			log.Printf("* http://%s:%v/", ipaddr, port)
		}
	}
	log.Println("===========================================")
	return nil
}

func ServeHttp(cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	port := 5555

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/favicon.ico", http.StatusSeeOther)
	})
	http.Handle("/static/", http.FileServer(http.FS(content)))
	http.HandleFunc("/", mainPage)

	listenOn := fmt.Sprintf("0.0.0.0:%v", port)

	err := connectionBanner(port)
	if err != nil {
		return err
	}

	go func() {
		log.Println("Starting HTTP server", port)

		err = http.ListenAndServe(listenOn, nil)
		if err != nil {
			cancel()
		}
	}()

	<-c.Done()
	log.Println("Finishing up HTTP server...")
	return err
}
