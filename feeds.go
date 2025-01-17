package goinsta

import (
	"encoding/json"
	"fmt"
)

// Feed is the object for all feed endpoints.
type Feed struct {
	insta *Instagram
}

// newFeed creates new Feed structure
func newFeed(insta *Instagram) *Feed {
	return &Feed{
		insta: insta,
	}
}

// Feed search by locationID
func (feed *Feed) LocationID(locationID int64) (*FeedLocation, error) {
	insta := feed.insta
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlFeedLocationID, locationID),
			Query: map[string]string{
				"rank_token":     insta.rankToken,
				"ranked_content": "true",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	res := &FeedLocation{}
	err = json.Unmarshal(body, res)
	return res, err
}

// FeedLocation is the struct that fits the structure returned by instagram on LocationID search.
type FeedLocation struct {
	RankedItems         []Item   `json:"ranked_items"`
	Items               []Item   `json:"items"`
	NumResults          int      `json:"num_results"`
	NextID              string   `json:"next_max_id"`
	MoreAvailable       bool     `json:"more_available"`
	AutoLoadMoreEnabled bool     `json:"auto_load_more_enabled"`
	MediaCount          int      `json:"media_count"`
	Location            Location `json:"location"`
	Status              string   `json:"status"`
}

// Tags search by Tag in user Feed
//
// (sorry for returning FeedTag. See #FeedTag)
func (feed *Feed) Tags(tag string) (*FeedTag, error) {
	insta := feed.insta
	body, err := insta.sendRequest(
		&reqOptions{
			Endpoint: fmt.Sprintf(urlFeedTag, tag),
			Query: map[string]string{
				"rank_token":     insta.rankToken,
				"ranked_content": "true",
			},
		},
	)
	if err != nil {
		return nil, err
	}
	res := &FeedTag{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	res.name = tag
	res.insta = feed.insta
	res.setValues()

	return res, nil
}

// FeedTag is the struct that fits the structure returned by instagram on TagSearch.
type FeedTag struct {
	insta *Instagram
	err   error

	name string

	RankedItems         []Item     `json:"ranked_items"`
	Images              []Item     `json:"items"`
	NumResults          int        `json:"num_results"`
	NextID              string     `json:"next_max_id"`
	MoreAvailable       bool       `json:"more_available"`
	AutoLoadMoreEnabled bool       `json:"auto_load_more_enabled"`
	Story               StoryMedia `json:"story"`
	Status              string     `json:"status"`
}

func (ft *FeedTag) setValues() {
	for i := range ft.RankedItems {
		ft.RankedItems[i].media = &FeedMedia{
			insta:  ft.insta,
			NextID: ft.RankedItems[i].ID,
		}
	}

	for i := range ft.Images {
		ft.Images[i].media = &FeedMedia{
			insta:  ft.insta,
			NextID: ft.Images[i].ID,
		}
	}
}

// Next paginates over hashtag feed.
func (ft *FeedTag) Next() bool {
	if ft.err != nil {
		return false
	}

	insta := ft.insta
	name := ft.name
	body, err := insta.sendRequest(
		&reqOptions{
			Query: map[string]string{
				"max_id":     ft.NextID,
				"rank_token": insta.rankToken,
			},
			Endpoint: fmt.Sprintf(urlFeedTag, name),
		},
	)
	if err == nil {
		newFT := &FeedTag{}
		err = json.Unmarshal(body, newFT)
		if err == nil {
			*ft = *newFT
			ft.insta = insta
			ft.name = name
			if !ft.MoreAvailable {
				ft.err = ErrNoMore
			}
			ft.setValues()
			return true
		}
	}
	ft.err = err
	return false
}

//Error returns hashtag error
func (ft *FeedTag) Error() error {
	return ft.err
}
