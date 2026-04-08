package services

import (
	"encoding/json"
	"fmt"
	"os"
	"scrobbler/discord"
	"scrobbler/music"
	"time"
)

type DiscordService struct {
	client      *discord.Client
	active      bool
	lastTrackID string
	wasPlaying  bool
	lastStart   int64
}

func NewDiscordService(appId string) *DiscordService {
	s := &DiscordService{
		client: discord.NewClient(appId),
	}
	go s.maintainConnection()
	return s
}

func (s *DiscordService) maintainConnection() {
	for {
		if !s.active {
			if err := s.client.Connect(); err == nil {
				s.active = true
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func (s *DiscordService) Update(track music.TrackInfo, enabled bool) {
	if !s.active {
		return
	}

	if !track.IsPlaying || !enabled {
		if s.wasPlaying {
			s.Clear()
			s.wasPlaying = false
			s.lastTrackID = ""
		}
		return
	}

	now := time.Now().UnixMilli()
	startTimestamp := now - track.Position.Milliseconds()
	endTimestamp := startTimestamp + track.Duration.Milliseconds()

	diff := startTimestamp - s.lastStart
	if diff < 0 {
		diff = -diff
	}

	if track.Title != s.lastTrackID || track.IsPlaying != s.wasPlaying || diff > 5000 {
		s.client.SetActivity(discord.Activity{
			Details: track.Title,
			State:   track.Artist,
			Type:    2, // Listening
			Timestamps: &discord.Timestamps{
				Start: startTimestamp,
				End:   endTimestamp,
			},
			Assets: &discord.Assets{
				LargeImage: GetAlbumArt(track.Artist, track.Album, track.Title),
				LargeText:  track.Album + " ",
			},
		})
	}

	s.lastTrackID = track.Title
	s.wasPlaying = track.IsPlaying
	s.lastStart = startTimestamp
}

func (s *DiscordService) Clear() {
	if !s.active {
		return
	}
	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": nil,
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	data, _ := json.Marshal(payload)
	s.client.Send(1, data)
}
