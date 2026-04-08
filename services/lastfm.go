package services

import (
	"fmt"
	"scrobbler/lastfm"
	"scrobbler/music"
	"strconv"
	"time"
)

type ScrobbleManager struct {
	currentTrackID      string
	hasScrobbled        bool
	hasSentNowPlaying   bool
	lastCheckTime       int64
	accumulatedPlayTime time.Duration
}

var Manager = &ScrobbleManager{}

func UpdatedNowPlaying(trackinfo music.TrackInfo, sessionKey string) error {
	if (sessionKey == "" || trackinfo == music.TrackInfo{}) {
		return fmt.Errorf("missing required track info or session key")
	}

	params := map[string]string{
		"method": "track.updateNowPlaying",
		"artist": trackinfo.Artist,
		"track":  trackinfo.Title,
		"sk":     sessionKey,
	}

	_, err := lastfm.PostLastfm(params)
	if err != nil {
		fmt.Println("Failed to update Now Playing:", err)
		return err
	}

	fmt.Printf("Last.fm: Now Playing -> %s\n", trackinfo.Title)
	return nil

}

func ScrobbleTrack(track music.TrackInfo, timestamp int64, sessionKey string) error {
	if sessionKey == "" || track.Title == "" {
		return nil
	}

	fmt.Printf("Last.fm: Scrobbled -> %s\n", track.Title)

	params := map[string]string{
		"method":    "track.scrobble",
		"artist":    track.Artist,
		"track":     track.Title,
		"album":     track.Album,
		"timestamp": strconv.FormatInt(timestamp, 10),
		"sk":        sessionKey,
	}

	_, err := lastfm.PostLastfm(params)
	return err
}

func (m *ScrobbleManager) ProcessScrobble(track music.TrackInfo, session string) {
	uniqueID := track.Artist + " - " + track.Title
	now := time.Now().Unix()

	if uniqueID != m.currentTrackID {
		m.currentTrackID = uniqueID
		m.hasScrobbled = false
		m.hasSentNowPlaying = false
		m.accumulatedPlayTime = 0
		m.lastCheckTime = now
	}

	delta := now - m.lastCheckTime
	m.lastCheckTime = now
	if track.IsPlaying {
		m.accumulatedPlayTime += time.Duration(delta) * time.Second
	}

	if track.IsPlaying && !m.hasSentNowPlaying && session != "" {
		m.hasSentNowPlaying = true
		go UpdatedNowPlaying(track, session)
	}

	if m.hasScrobbled || !track.IsPlaying || track.Duration < (30*time.Second) {
		return
	}

	limit := track.Duration / 2
	if limit > (4 * time.Minute) {
		limit = 4 * time.Minute
	}

	if m.accumulatedPlayTime >= limit || track.Position >= limit {
		m.hasScrobbled = true
		startTime := now - int64(m.accumulatedPlayTime.Seconds())
		go ScrobbleTrack(track, startTime, session)
	}
}
