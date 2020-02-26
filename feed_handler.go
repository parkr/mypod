package mypod

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/jbub/podcasts"
	"github.com/technoweenie/grohl"
)

func NewFeedHandler(dir string, logger *grohl.Context) http.Handler {
	return &FeedHandler{
		dir:    dir,
		logger: logger,
	}
}

type FeedHandler struct {
	dir    string
	logger *grohl.Context
}

func (h *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	feed, err := h.GetFeed()
	if err != nil {
		http.Error(w, "error generating feed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := feed.Write(w); err != nil {
		http.Error(w, "error writing feed: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) GetFeed() (*podcasts.Feed, error) {
	conf, err := ReadConfig(filepath.Join(h.dir, "podcast.json"))
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(conf.BaseURL)
	if err != nil {
		return nil, err
	}

	// initialize the podcast
	podcast := &podcasts.Podcast{
		Title:       conf.Title,
		Description: conf.Description,
		Language:    conf.Language,
		Link:        conf.Link,
		Copyright:   conf.Copyright,
	}

	// add items
	items, err := h.ReadPodcastEpisodes()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		itemPath := item.Enclosure.URL
		itemURL := &url.URL{
			Scheme: baseURL.Scheme,
			Opaque: baseURL.Opaque,
			User:   baseURL.User,
			Host:   baseURL.Host,
			Path:   "/files/" + itemPath,
		}
		item.Enclosure.URL = itemURL.String()
		podcast.AddItem(item)
	}

	// build feed
	return podcast.Feed(
		podcasts.Author(conf.Author),
		podcasts.NewFeedURL(conf.BaseURL+"/feed.xml"),
		podcasts.Subtitle(conf.Subtitle),
		podcasts.Summary(conf.Summary),
		podcasts.Owner(conf.Owner.Name, conf.Owner.Email),
		podcasts.Image(conf.Image),
	)
}

func (h *FeedHandler) ReadPodcastEpisodes() ([]*podcasts.Item, error) {
	items := []*podcasts.Item{}
	walkErr := filepath.Walk(h.dir+"/files", func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileLocation, err := filepath.Rel(h.dir, filePath)
		if err != nil {
			return err
		}

		mime, err := mimetype.DetectFile(filePath)
		if err != nil {
			return err
		}

		items = append(items, &podcasts.Item{
			Title:   titleize(fileLocation),
			PubDate: &podcasts.PubDate{Time: info.ModTime()},
			GUID:    hash(filePath),
			Enclosure: &podcasts.Enclosure{
				URL:  filepath.Base(fileLocation),
				Type: mime.String(),
			},
		})
		return nil
	})
	return items, walkErr
}

func titleize(fileLocation string) string {
	title := fileLocation
	// Give me the base name
	title = filepath.Base(title)
	// Remove extension
	title = title[0 : len(title)-len(filepath.Ext(title))]
	// Replace underscores with spaces
	title = strings.Replace(title, "_", " ", -1)
	// Capitalize
	title = strings.Title(title)
	return title
}

func hash(str string) string {
	sum := sha256.Sum256([]byte(str))
	return hex.EncodeToString(sum[:])
}
