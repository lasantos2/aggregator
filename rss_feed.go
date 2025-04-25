package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(feedUrl)
	if err != nil {
	
		return &RSSFeed{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	
	if err != nil {
	
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent","gator")
	resp, err = client.Do(req)

	if err != nil {
		return &RSSFeed{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	rssFeed := RSSFeed{}

	xml.Unmarshal(body, &rssFeed)
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for ind, _ := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[ind].Title = html.UnescapeString(rssFeed.Channel.Item[ind].Title)
		rssFeed.Channel.Item[ind].Description = html.UnescapeString(rssFeed.Channel.Item[ind].Description)
	}

	return &rssFeed, nil
}