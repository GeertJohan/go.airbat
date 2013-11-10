package main

import (
	"fmt"
	"github.com/GeertJohan/go.airbat"
	"github.com/GeertJohan/go.airbrake"
	"github.com/jessevdk/go-flags"
	"html/template"
	"log"
	"net/http"
	"os"
)

const airbrakeNoticeURL = "http://airbrake.io/locate/%d"

var (
	tmplError    *template.Template
	tmplRedirect *template.Template
)

func init() {
	var err error
	tmplError, err = template.New("error").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<title>Airb.at - This is not the error you are looking for</title>
		<meta name="description" content="Airb.at short urls for airbrake.io">
		<meta name="author" content="Geert-Johan Riemer">
	</head>
	<body>
		<center>
			<h3>This is not the error you are looking for.</h3>
			<p>
				An error occurred. The given URL seems to be inavalid.<br/>
				{{.Error}}
			</p>
		</center>
	</body>
</html>`)
	if err != nil {
		log.Printf("error parsing error template: %s\n", err)
		os.Exit(1)
	}

	tmplRedirect, err = template.New("redirect").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<title>Airb.at - You are being redirected</title>
		<meta name="description" content="Airb.at short urls for airbrake.io">
		<meta name="author" content="Geert-Johan Riemer">
	</head>
	<body>
		<center>
			<h3>Redirecting..</h3>
			<p>
				You are being redirected to: <a href="{{.URL}}" >{{.URL}}</a>
			</p>
		</center>
	</body>
</html>`)
	if err != nil {
		log.Printf("error parsing redirect template: %s\n", err)
		os.Exit(1)
	}
}

type dataError struct {
	Error string
}

type dataRedirect struct {
	URL template.URL
}

var options struct {
	HTTPPort    string `long:"port" default:"8321"`
	AirID       string `long:"airid"`
	AirKey      string `long:"airkey"`
	Environment string `long:"environment"`
}

func main() {
	// parse flags
	_, err := flags.Parse(&options)
	if err != nil {
		fmt.Printf("error parsing flags: %s\n", err)
		os.Exit(1)
	}

	if options.AirID == "" || options.AirKey == "" || options.Environment == "" {
		fmt.Println("require both --airid and --airkey and --environment flags")
		os.Exit(1)
	}

	brake := airbrake.NewBrake(options.AirID, options.AirKey, options.Environment, &airbrake.Config{
		URLService: airbrake.URLService_Airbat,
	})

	// set http handler on root request uri
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get code from RequestURI
		code := r.RequestURI[1:]
		if len(code) == 0 {
			http.Redirect(w, r, "https://github.com/GeertJohan/go.airbat", http.StatusTemporaryRedirect)
			return
		}

		// convert code to notice id
		id, err := airbat.AirbatCodeToUint(code)
		if err != nil {
			//++ TODO: should this be reported to airbrake? This can be just a wrong url (user-edit gone wrong)
			brake.Errorf("conversion error", "could not convert to uint: %s", err)

			// write error template
			data := dataError{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			err = tmplError.Execute(w, data)
			if err != nil {
				brake.Errorf("template error", "error executing error template: %s", err)
				return
			}
			return
		}

		// create airbrake url
		urlStr := fmt.Sprintf(airbrakeNoticeURL, id)

		// write redirect header
		w.Header().Set("Location", urlStr)
		w.WriteHeader(http.StatusTemporaryRedirect)

		// RFC2616 recommends that a short note "SHOULD" be included in the
		// response because older user agents may not understand 301/307.
		// Shouldn't send the response for POST or HEAD; that leaves GET.
		if r.Method == "GET" {
			// setup data for template
			data := dataRedirect{
				URL: template.URL(urlStr),
			}
			// execute template
			err = tmplRedirect.Execute(w, data)
			if err != nil {
				brake.Errorf("template error", "error executing redirect template: %s", err)
			}
		}
	})

	// start http server in goroutine
	go func() {
		err := http.ListenAndServe(":"+options.HTTPPort, nil)
		if err != nil {
			log.Printf("error listening on port %s: %s\n", options.HTTPPort, err)
			os.Exit(1)
		}
	}()

	// wait forever
	select {}
}
