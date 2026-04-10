package rss

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/store"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    AtomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type AtomEntry struct {
	Title   string   `xml:"title"`
	Link    AtomLink `xml:"link"`
	ID      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Summary string   `xml:"summary"`
	Content AtomContent `xml:"content"`
}

type AtomContent struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

func GenerateRSS(articles []store.Article, baseURL string) ([]byte, error) {
	items := make([]Item, 0, len(articles))
	for _, a := range articles {
		link := a.Link
		if link == "" {
			link = fmt.Sprintf("%s/articles/%s", baseURL, a.ID)
		}
		items = append(items, Item{
			Title:       a.Title,
			Link:        link,
			Description: a.Description,
			Content:     a.Markdown,
			PubDate:     time.UnixMilli(a.CreatedAt).UTC().Format(time.RFC1123Z),
			GUID:        fmt.Sprintf("xinzhi-%s", a.XinzhiID),
		})
	}

	feed := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         "参考答案阅览室 - RSS Feed",
			Link:          baseURL,
			Description:   "参考答案阅览室的订阅文章，由 ReferenceAnswerRSS 自动生成",
			Language:      "zh-CN",
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			Items:         items,
		},
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), output...), nil
}

func GenerateAtom(articles []store.Article, baseURL string) ([]byte, error) {
	entries := make([]AtomEntry, 0, len(articles))
	for _, a := range articles {
		link := a.Link
		if link == "" {
			link = fmt.Sprintf("%s/articles/%s", baseURL, a.ID)
		}
		entries = append(entries, AtomEntry{
			Title:   a.Title,
			Link:    AtomLink{Href: link, Rel: "alternate", Type: "text/html"},
			ID:      fmt.Sprintf("xinzhi-%s", a.XinzhiID),
			Updated: time.UnixMilli(a.CreatedAt).UTC().Format(time.RFC3339),
			Summary: a.Description,
			Content: AtomContent{Type: "text", Content: a.Markdown},
		})
	}

	feed := AtomFeed{
		XMLNS:   "http://www.w3.org/2005/Atom",
		Title:   "参考答案阅览室 - Atom Feed",
		Link:    AtomLink{Href: baseURL + "/feed.atom", Rel: "self", Type: "application/atom+xml"},
		Updated: time.Now().UTC().Format(time.RFC3339),
		ID:      baseURL,
		Entries: entries,
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), output...), nil
}
