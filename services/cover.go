package services

import (
	"encoding/json"
	"fmt"
	"gala/config"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var (
	cacheMu   sync.RWMutex
	lastKey   string
	cachedURL string
	lastFetch time.Time
)

var reSanitize = regexp.MustCompile(`(?i)\s*[\(\[].*?(deluxe|remaster|single|edition|version|explicit|anniversary).*?[\)\]]`)

func GetAlbumArt(artist, album, title string) string {
	currentKey := fmt.Sprintf("%s|%s", artist, album)
	if album == "" {
		currentKey = fmt.Sprintf("%s|%s", artist, title)
	}

	cacheMu.RLock()
	if currentKey == lastKey && cachedURL != "" {
		defer cacheMu.RUnlock()
		return cachedURL
	}
	cacheMu.RUnlock()

	conf, _ := config.LoadConfig()
	apiKey := conf.GetApiKey()
	if apiKey == "" {
		return "applemusic"
	}

	if time.Since(lastFetch) < 2*time.Second {
		return "applemusic"
	}

	// 5. Build Last.fm API URL
	v := url.Values{}
	v.Set("method", "album.getinfo")
	v.Set("api_key", apiKey)
	v.Set("artist", artist)
	v.Set("album", album)
	v.Set("format", "json")
	if album == "" {
		v.Set("method", "track.getInfo")
		v.Set("track", title)
	}

	apiURL := "https://ws.audioscrobbler.com/2.0/?" + v.Encode()

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(apiURL)
	lastFetch = time.Now()
	if err != nil {
		return "applemusic"
	}
	defer resp.Body.Close()

	var result struct {
		Album struct {
			Image []struct {
				Text string `json:"#text"`
				Size string `json:"size"`
			} `json:"image"`
		} `json:"album"`
		Track struct {
			Album struct {
				Image []struct {
					Text string `json:"#text"`
					Size string `json:"size"`
				} `json:"image"`
			} `json:"album"`
		} `json:"track"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "applemusic"
	}

	var foundURL string
	images := result.Album.Image
	if len(images) == 0 {
		images = result.Track.Album.Image
	}

	for _, img := range images {
		if img.Size == "extralarge" || img.Size == "mega" {
			foundURL = img.Text
		}
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()
	lastKey = currentKey
	if foundURL == "" {
		cachedURL = "applemusic"
	} else {
		cachedURL = foundURL
	}

	return cachedURL
}
