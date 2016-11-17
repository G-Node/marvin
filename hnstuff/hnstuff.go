package hnstuff

import (
	"log"

	"github.com/SlyMarbo/rss"
)

type HnFeed struct {
	Name   string
	HNFeed *rss.Feed
}

func NewHnFeed() HnFeed {
	feed, err := rss.Fetch("http://hnrss.org/newest?points=300")
	if err != nil {
		log.Printf("Could nor fetch HnFedd: %v", err)
		return HnFeed{}
	}

	for _, item := range feed.Items {
		item.Read = true
	}
	return HnFeed{"feed", feed}
}

func (f *HnFeed) GetNewStories() []*rss.Item {
	newItems := make([]*rss.Item, 0)
	for _, item := range f.HNFeed.Items {
		if !item.Read {
			item.Read = true
			newItems = append(newItems, item)
		}
	}
	return newItems
}
