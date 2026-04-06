package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"scrobbler/config"
	"scrobbler/lastfm"
	"scrobbler/music"
	"scrobbler/services"
	"time"

	"github.com/getlantern/systray"
)

var iconData []byte

var LASTFM_API_KEY = ""
var LASTFM_SECRET = ""

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	//systray.SetIcon(iconData)
	systray.SetTitle("Scrobbler")
	systray.SetTooltip("wa")

	//Tray menu
	mStatus := systray.AddMenuItem("Status: Idle", "Current scrobbling status")
	mStatus.Disable()
	systray.AddSeparator()
	mAuth := systray.AddMenuItem("Login to last.fm", "Authenticate via browser")
	mLogout := systray.AddMenuItem("Logout", "Close the current lastfm session")
	mLogout.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Close the app")

	conf, err := config.LoadConfig()

	if err == nil && conf.LastFmApiKey == "" || conf.LastFmSecret == "" {
		config.SaveConfig(conf)
		exec.Command("cmd", "/c", "start", config.GetConfigPath()).Run()

	}

	if err == nil && conf.LastFmSessionKey != "" {
		mAuth.SetTitle("Logged in as: " + conf.LastFmUsername)
		mAuth.Disable()
		mLogout.Enable()
	}

	for conf.LastFmApiKey == "" || conf.LastFmSecret == "" {
		time.Sleep(2 * time.Second)
		conf, _ = config.LoadConfig()
		// we hold the program hostage until the key and secret are provided on the config file
	}

	LASTFM_API_KEY = conf.LastFmApiKey
	LASTFM_SECRET = conf.LastFmSecret

	go func() {
		for {
			select {
			case <-mAuth.ClickedCh:
				attempLogin(mAuth, mLogout)

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-mLogout.ClickedCh:
				logout(mAuth, mLogout)
			}
		}
	}()

	go handleTrackEvent(conf, mStatus)

	go runScrobblerBackground(conf)

	go music.MonitorMusic(mStatus, conf)
}

func runScrobblerBackground(conf *config.Config) {
	for {
		track := music.GetTrack()

		services.Manager.ProcessScrobble(track, conf.LastFmSessionKey)

		time.Sleep(1 * time.Second)
	}
}

func handleTrackEvent(conf *config.Config, mStatus *systray.MenuItem) {
	for track := range music.TrackEventChan {
		if track.IsPlaying {
			mStatus.SetTitle(track.Title)
		} else {
			fmt.Printf("⏸️ Event: Paused %s\n", track.Title)
		}
	}
}

func attempLogin(mAuth *systray.MenuItem, mLogout *systray.MenuItem) {
	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	token, _ := lastfm.StartAuthServer(LASTFM_API_KEY)
	session, user, err := lastfm.FetchSessionKey(token, LASTFM_API_KEY, LASTFM_SECRET)

	if err != nil {
		fmt.Errorf("something wrong happend : %s", err)
	}

	conf.SetSessionKey(session)
	conf.SetUsername(user)

	mAuth.SetTitle("Logged in as: " + user)
	mAuth.Disable()
	mLogout.Enable()

}

func logout(mAuth *systray.MenuItem, mLogout *systray.MenuItem) {

	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	conf.SetSessionKey("")
	conf.SetUsername("")

	mAuth.SetTitle("Login to last.fm")
	mAuth.Enable()
	mLogout.Disable()

}

func onExit() {
	fmt.Println("Closing down")
}
