package mypod

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
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
	items, err := h.ReadPodcastEpisodes(conf)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		// Complete podcast episode URL
		itemPath := item.Enclosure.URL
		itemURL := &url.URL{
			Scheme: baseURL.Scheme,
			Opaque: baseURL.Opaque,
			User:   baseURL.User,
			Host:   baseURL.Host,
			Path:   "/files/" + itemPath,
		}
		item.Enclosure.URL = itemURL.String()

		// Complete podcast episode image URL
		imagePath := item.Image.Href
		imageURL := &url.URL{
			Scheme: baseURL.Scheme,
			Opaque: baseURL.Opaque,
			User:   baseURL.User,
			Host:   baseURL.Host,
			Path:   imagePath,
		}
		item.Image.Href = imageURL.String()

		podcast.AddItem(item)
	}

	imageURL := &url.URL{
		Scheme: baseURL.Scheme,
		Opaque: baseURL.Opaque,
		User:   baseURL.User,
		Host:   baseURL.Host,
		conf.Image,
	}

	// build feed
	return podcast.Feed(
		podcasts.Author(conf.Author),
		podcasts.NewFeedURL(conf.BaseURL+"/feed.xml"),
		podcasts.Subtitle(conf.Subtitle),
		podcasts.Summary(conf.Summary),
		podcasts.Owner(conf.Owner.Name, conf.Owner.Email),
		podcasts.Image(imageURL.String()),
		setCategories(conf),
		setExplicit(conf.Explicit),
	)
}

func setCategories(conf Config) func(f *podcasts.Feed) error {
	return func(f *podcasts.Feed) error {
		f.Channel.Categories = []*podcasts.ItunesCategory{}
		for _, category := range conf.Categories {
			f.Channel.Categories = append(f.Channel.Categories, &podcasts.ItunesCategory{
				Text: category,
			})
		}
		return nil
	}
}

func setExplicit(explicit bool) func(f *podcasts.Feed) error {
	return func(f *podcasts.Feed) error {
		if explicit {
			f.Channel.Explicit = podcasts.ValueYes
		} else {
			f.Channel.Explicit = "no"
		}
		return nil
	}
}

func (h *FeedHandler) ReadPodcastEpisodes(conf Config) ([]*podcasts.Item, error) {
	imagePaths, err := listImagePaths(h.dir + "/images")
	if err != nil {
		grohl.Log(grohl.Data{
			"msg":   "listing thumbnail images",
			"error": err.Error(),
		})
	}

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

		item := &podcasts.Item{
			Title:    titleize(fileLocation),
			PubDate:  podcasts.NewPubDate(info.ModTime()),
			GUID:     hash(filePath + info.ModTime().String()),
			Author:   conf.Author,
			Subtitle: "A subtitle for this episode",
			Summary:  &podcasts.ItunesSummary{Value: "A summary for this episode"},
			Enclosure: &podcasts.Enclosure{
				URL:    filepath.Base(fileLocation),
				Length: strconv.FormatInt(info.Size(), 10),
			},
			Explicit: "no",
		}

		if thumbnailPath := episodeThumbnailPath(imagePaths, filePath); thumbnailPath != "" {
			item.Image = &podcasts.ItunesImage{Href: thumbnailPath}
		} else {
			item.Image = &podcasts.ItunesImage{Href: conf.Image}
		}

		if mime, err := mimetype.DetectFile(filePath); err == nil {
			item.Enclosure.Type = mime.String()
		}

		if metadata, err := readMetadata(filePath); err == nil {
			if title := metadata.Title(); title != "" {
				item.Title = title
			}
			if author := metadata.Artist(); author != "" {
				item.Author = author
			}
			if comment := metadata.Comment(); comment != "" {
				item.Summary = &podcasts.ItunesSummary{Value: comment}
			}
		}

		if duration, err := readDuration(filePath); err == nil {
			item.Duration = podcasts.NewDuration(duration)
		}

		items = append(items, item)
		return nil
	})

	// Sort by pubDate DESC (newer first, older last).
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].PubDate.Time.After(items[j].PubDate.Time)
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

func listImagePaths(imagesDir string) ([]string, error) {
	infos, err := ioutil.ReadDir(imagesDir)
	if err != nil {
		return []string{}, err
	}
	thumbnails := []string{}
	for _, info := range infos {
		if !info.IsDir() {
			thumbnails = append(thumbnails, info.Name())
		}
	}
	return thumbnails, nil
}

func episodeThumbnailPath(imagePaths []string, filePath string) string {
	// Try the exact episode name
	episodeName := filepath.Base(filePath)
	episodeName = episodeName[0 : len(episodeName)-len(filepath.Ext(episodeName))]

	for _, thumbnailFileName := range imagePaths {
		if strings.HasPrefix(thumbnailFileName, episodeName) {
			return "/images/" + thumbnailFileName
		}
	}

	// Try removing the video key
	if idx := strings.LastIndex(episodeName, "-"); idx > 0 {
		episodeName = episodeName[0:idx]
		for _, thumbnailFileName := range imagePaths {
			if strings.HasPrefix(thumbnailFileName, episodeName) {
				return "/images/" + thumbnailFileName
			}
		}
	}

	return ""
}

func readMetadata(filePath string) (tag.Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return tag.ReadFrom(f)
}

func readDuration(filePath string) (time.Duration, error) {
	var buf bytes.Buffer
	cmd := exec.Command("ffprobe", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", "-i", filePath)
	cmd.Stdout = &buf
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	// ffprobe responds with the total number of seconds.
	// itunes:duration should be in HH:MM:SS format, but since the library
	// requires a time.Duration, the best we can do is list the duration in seconds.
	// When the XML marshaling process occurs, we want to see the raw seconds value,
	// so we thus return the highest precision value we can for Duration, namely ns.
	return time.ParseDuration(strings.TrimSpace(buf.String()) + "ns")
}
