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
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
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

func cleanupUsername(username string) string {
	var re = regexp.MustCompile(`[^A-Za-z0-9]+`)
	username = re.ReplaceAllString(username, `_`)

	username = strings.Replace(username, " ", "_", -1)

	return username
}

func ServeHttp(command func(string), cancel context.CancelFunc, c context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	httpServerPort := 5555

	mainPage := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			host := r.Host
			hostname, _, _ := net.SplitHostPort(host)

			r.ParseForm()
			username := cleanupUsername(r.Form.Get("username"))

			connectionstring := []string{username, "", hostname, "25565"}
			c, _ := json.Marshal(connectionstring)

			frontpagetemplate.Execute(w, RenderFlags{ShowCanvas: true, ConnectionString: string(c)})

			go func() {
				time.Sleep(2 * time.Second)
				command(fmt.Sprintf("Oh hi, %s", username))
			}()
			return
		}

		w.WriteHeader(200)
		w.Header().Add("content-type", "text/html")
		frontpagetemplate.Execute(w, RenderFlags{ShowCanvas: false})
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/favicon.ico", http.StatusSeeOther)
	})
	http.Handle("/static/", http.FileServer(http.FS(content)))
	http.HandleFunc("/", mainPage)

	listenOn := fmt.Sprintf("0.0.0.0:%v", httpServerPort)

	err := connectionBanner(httpServerPort)
	if err != nil {
		return err
	}

	go func() {
		log.Println("Starting HTTP server", httpServerPort)

		err = http.ListenAndServe(listenOn, nil)
		if err != nil {
			cancel()
		}
	}()

	<-c.Done()
	log.Println("Finishing up HTTP server...")
	return err
}
