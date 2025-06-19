package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sternelee/OpenList-workers/v3/internal/model"
)

// RSS XML 结构定义
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Atom struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	Updated string      `xml:"updated"`
	Author  AtomAuthor  `xml:"author"`
	ID      string      `xml:"id"`
	Entries []AtomEntry `xml:"entry"`
}

type Channel struct {
	Title         string `xml:"title"`
	Description   string `xml:"description"`
	Link          string `xml:"link"`
	Language      string `xml:"language"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	GUID        string    `xml:"guid"`
	PubDate     string    `xml:"pubDate"`
	Author      string    `xml:"author"`
	Category    string    `xml:"category"`
	Enclosure   Enclosure `xml:"enclosure"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title   string     `xml:"title"`
	Summary string     `xml:"summary"`
	Content string     `xml:"content"`
	Link    []AtomLink `xml:"link"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Author  AtomAuthor `xml:"author"`
}

// FeedParser RSS/Atom feed 解析器
type FeedParser struct {
	client *http.Client
}

func NewFeedParser() *FeedParser {
	return &FeedParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseFeed 解析 RSS/Atom feed
func (p *FeedParser) ParseFeed(url string) (*model.RSSFeed, []model.RSSArticle, error) {
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 尝试解析 RSS
	if articles, feedInfo, err := p.parseRSS(body); err == nil {
		feedInfo.URL = url
		return feedInfo, articles, nil
	}

	// 尝试解析 Atom
	if articles, feedInfo, err := p.parseAtom(body); err == nil {
		feedInfo.URL = url
		return feedInfo, articles, nil
	}

	return nil, nil, fmt.Errorf("unsupported feed format")
}

func (p *FeedParser) parseRSS(data []byte) ([]model.RSSArticle, *model.RSSFeed, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, nil, err
	}

	feedInfo := &model.RSSFeed{
		Name: rss.Channel.Title,
	}

	var articles []model.RSSArticle
	for _, item := range rss.Channel.Items {
		article := model.RSSArticle{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			GUID:        item.GUID,
			Author:      item.Author,
			Category:    item.Category,
		}

		// 解析发布时间
		if pubDate, err := p.parseTime(item.PubDate); err == nil {
			article.PubDate = pubDate
		}

		// 解析种子和磁力链接
		article.TorrentURL, article.MagnetLink = p.extractLinks(item.Description, item.Link, item.Enclosure.URL)

		// 解析文件大小
		if item.Enclosure.Length != "" {
			if size, err := strconv.ParseInt(item.Enclosure.Length, 10, 64); err == nil {
				article.FileSize = size
			}
		}

		// 解析种子信息
		article.Seeders, article.Leechers = p.extractSeedInfo(item.Description)

		articles = append(articles, article)
	}

	return articles, feedInfo, nil
}

func (p *FeedParser) parseAtom(data []byte) ([]model.RSSArticle, *model.RSSFeed, error) {
	var atom Atom
	if err := xml.Unmarshal(data, &atom); err != nil {
		return nil, nil, err
	}

	feedInfo := &model.RSSFeed{
		Name: atom.Title,
	}

	var articles []model.RSSArticle
	for _, entry := range atom.Entries {
		article := model.RSSArticle{
			Title:       entry.Title,
			Description: entry.Summary,
			GUID:        entry.ID,
			Author:      entry.Author.Name,
		}

		if entry.Content != "" {
			article.Description = entry.Content
		}

		// 找到文章链接
		for _, link := range entry.Link {
			if link.Rel == "alternate" || link.Rel == "" {
				article.URL = link.Href
				break
			}
		}

		// 解析更新时间
		if updated, err := p.parseTime(entry.Updated); err == nil {
			article.PubDate = updated
		}

		// 解析种子和磁力链接
		article.TorrentURL, article.MagnetLink = p.extractLinks(article.Description, article.URL, "")

		articles = append(articles, article)
	}

	return articles, feedInfo, nil
}

func (p *FeedParser) parseTime(timeStr string) (time.Time, error) {
	// 支持多种时间格式
	formats := []string{
		time.RFC822,
		time.RFC822Z,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", timeStr)
}

func (p *FeedParser) extractLinks(description, itemURL, enclosureURL string) (torrentURL, magnetLink string) {
	// 提取磁力链接
	magnetRegex := regexp.MustCompile(`magnet:\?[^"'\s<>]+`)
	if matches := magnetRegex.FindString(description); matches != "" {
		magnetLink = matches
	}

	// 提取种子链接
	torrentRegex := regexp.MustCompile(`https?://[^"'\s<>]+\.torrent(?:\?[^"'\s<>]*)?`)
	if matches := torrentRegex.FindString(description); matches != "" {
		torrentURL = matches
	} else if strings.HasSuffix(itemURL, ".torrent") {
		torrentURL = itemURL
	} else if strings.HasSuffix(enclosureURL, ".torrent") {
		torrentURL = enclosureURL
	}

	return torrentURL, magnetLink
}

func (p *FeedParser) extractSeedInfo(description string) (seeders, leechers int) {
	// 尝试从描述中提取种子和下载者信息
	seedRegex := regexp.MustCompile(`(?i)seed(?:er)?s?[:\s]*(\d+)`)
	leechRegex := regexp.MustCompile(`(?i)leech(?:er)?s?[:\s]*(\d+)`)

	if matches := seedRegex.FindStringSubmatch(description); len(matches) > 1 {
		if s, err := strconv.Atoi(matches[1]); err == nil {
			seeders = s
		}
	}

	if matches := leechRegex.FindStringSubmatch(description); len(matches) > 1 {
		if l, err := strconv.Atoi(matches[1]); err == nil {
			leechers = l
		}
	}

	return seeders, leechers
}