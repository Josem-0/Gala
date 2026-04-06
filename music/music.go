package music

import (
	"bytes"
	"fmt"
	"runtime"
	"scrobbler/config"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
)

type TrackInfo struct {
	Title     string
	Artist    string
	Album     string
	AppID     string
	Position  time.Duration
	Duration  time.Duration
	IsPlaying bool
}

type trackState struct {
	mu   sync.RWMutex
	info TrackInfo
}

var TrackEventChan = make(chan TrackInfo, 10)
var current = &trackState{}

func UpdateTrack(title, artist, album, appID string, pos, dur time.Duration, isPlaying bool) {
	current.mu.Lock()
	defer current.mu.Unlock()

	current.info.Title = title
	current.info.Artist = artist
	current.info.Album = album
	current.info.AppID = appID
	current.info.Position = pos
	current.info.Duration = dur
	current.info.IsPlaying = isPlaying
}

func GetTrack() TrackInfo {
	current.mu.RLock() // Read lock
	defer current.mu.RUnlock()

	return current.info
}

func MonitorMusic(statusItem *systray.MenuItem, conf *config.Config) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	mediaDLL := syscall.NewLazyDLL("media.dll")
	getMediaProc := mediaDLL.NewProc("GetCurrentMedia")
	var playStatus int32

	var lastTitle, lastArtist string
	var lastPlayState bool

	for {
		titleBuf := make([]byte, 256)
		artistBuf := make([]byte, 256)
		albumBuf := make([]byte, 256)
		appBuf := make([]byte, 256)

		var positionTicks, durationTicks int64

		getMediaProc.Call(
			uintptr(unsafe.Pointer(&titleBuf[0])), uintptr(len(titleBuf)),
			uintptr(unsafe.Pointer(&artistBuf[0])), uintptr(len(artistBuf)),
			uintptr(unsafe.Pointer(&albumBuf[0])), uintptr(len(albumBuf)),
			uintptr(unsafe.Pointer(&appBuf[0])), uintptr(len(appBuf)),
			uintptr(unsafe.Pointer(&positionTicks)),
			uintptr(unsafe.Pointer(&durationTicks)),
			uintptr(unsafe.Pointer(&playStatus)),
		)

		Title := string(titleBuf[:bytes.IndexByte(titleBuf, 0)])
		Artist := string(artistBuf[:bytes.IndexByte(artistBuf, 0)])
		Album := string(albumBuf[:bytes.IndexByte(albumBuf, 0)])
		AppID := string(appBuf[:bytes.IndexByte(appBuf, 0)])

		//apple music fix
		Artist, Album = parseAppleMusicArtistAlbum(Artist, Album)

		Position := time.Duration(positionTicks * 100)
		Duration := time.Duration(durationTicks * 100)
		IsPlaying := playStatus == 4

		songChanged := Title != lastTitle || Artist != lastArtist
		statusChanged := IsPlaying != lastPlayState

		UpdateTrack(Title, Artist, Album, AppID, Position, Duration, IsPlaying)

		if songChanged || statusChanged {
			lastTitle = Title
			lastArtist = Artist
			lastPlayState = IsPlaying

			UpdateTrack(Title, Artist, Album, AppID, Position, Duration, IsPlaying)

			if Title != "" && isAllowedApp(AppID) {
				TrackEventChan <- GetTrack()
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func isAllowedApp(appId string) bool {
	appLower := strings.ToLower(appId)

	allowed := []string{
		"applemusicwin",
	}

	for _, allowedApp := range allowed {
		if strings.Contains(appLower, allowedApp) {
			return true
		}
	}
	return false
}

func formatTime(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// parseAppleMusicArtistAlbum splits the artist and album if the artist string contains a separator.
func parseAppleMusicArtistAlbum(artist, album string) (string, string) {
	separator := ""
	if strings.Contains(artist, " — ") {
		separator = " — "
	} else if strings.Contains(artist, " - ") {
		separator = " - "
	}

	if separator != "" {
		parts := strings.SplitN(artist, separator, 2)
		artist = strings.TrimSpace(parts[0])
		if album == "" && len(parts) > 1 {
			album = strings.TrimSpace(parts[1])
		}
	}
	return artist, album
}
