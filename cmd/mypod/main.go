package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/parkr/mypod"
	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

func getMailgunService() radar.MailgunService {
	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		radar.Println("unable to fetch mailgun from env:", err)
	}
	return radar.NewMailgunService(mg, os.Getenv("MG_FROM_EMAIL"))
}

func readAdditionalYtdlArgs(storage string) []string {
	ytdlArgsFile := storage + "/yt-dl-args.json"
	if _, err := os.Stat(ytdlArgsFile); os.IsNotExist(err) {
		return []string{}
	}

	ytdlArgsFileHandler, err := os.Open(ytdlArgsFile)
	if err != nil {
		radar.Println("unable to read yt-dl-args.json:", err)
		return []string{}
	}
	ytdlArgs := []string{}
	err = json.NewDecoder(ytdlArgsFileHandler).Decode(&ytdlArgs)
	if err != nil {
		radar.Println("unable to decode yt-dl-args.json:", err)
		return []string{}
	}
	return ytdlArgs
}

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":5312", "The IP/PORT to bind this server to.")
	var debug bool
	flag.BoolVar(&debug, "debug", os.Getenv("DEBUG") == "", "Whether to print debugging messages.")
	var storage string
	flag.StringVar(&storage, "storage", "/storage", "Where to store and serve the files")
	flag.Parse()

	grohl.SetLogger(grohl.NewIoLogger(os.Stderr))
	grohl.SetStatter(nil, 0, "")

	mux := http.NewServeMux()
	downloadService := mypod.NewDownloadService(storage, readAdditionalYtdlArgs(storage))

	radarGeneratedChan := make(chan bool)
	emailHandler := radar.NewEmailHandler(
		downloadService,
		getMailgunService(),
		strings.Split(os.Getenv("MYPOD_ALLOWED_SENDERS"), ","), // Allowed senders (email addresses)
		debug,              // Whether in debug mode
		radarGeneratedChan, // Channel to signal when to generate a new radar issue, not used.
	)
	mux.Handle("/emails", emailHandler)
	mux.Handle("/email", emailHandler)

	feedHandler := mypod.NewFeedHandler(storage, grohl.CurrentContext)
	mux.Handle("/feed.xml", feedHandler)

	mux.Handle("/files/", http.StripPrefix("/files", http.FileServer(http.Dir(storage+"/files"))))
	mux.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir(storage+"/images"))))
	mux.Handle("/", http.FileServer(http.Dir(storage+"/static")))

	go emailHandler.Start()

	handler := mypod.AdditionalLogContextHandler(mux)
	handler = radar.LoggingHandler(handler)

	radar.Println("Starting server on", binding)
	server := &http.Server{Addr: binding, Handler: handler}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		radar.Println("Closing database connection...")
		downloadService.Shutdown(ctx)
		emailHandler.Shutdown(ctx)
		_ = server.Shutdown(ctx)
		radar.Println("Done with graceful shutdown.")
	}()

	if err := server.ListenAndServe(); err != nil {
		radar.Println("error listening:", err)
	}
}
