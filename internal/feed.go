package blog

import (
	"encoding/xml"
	"net/http"
	"time"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func (app *App) handleRSS(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	posts := publicPosts(app.posts)
	app.mu.RUnlock()

	items := make([]rssItem, len(posts))
	for i, p := range posts {
		link := app.cfg.BaseURL + "/posts/" + p.Slug
		items[i] = rssItem{
			Title:       p.Title,
			Link:        link,
			Description: p.Description,
			PubDate:     p.Date.Format(time.RFC1123Z),
			GUID:        link,
		}
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       "thobiasn.dev",
			Link:        app.cfg.BaseURL,
			Description: "Personal blog by thobiasn",
			LastBuild:   time.Now().Format(time.RFC1123Z),
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(feed)
}
