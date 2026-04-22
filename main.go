package main

import (
	_ "embed"
	"fmt"
	"gala/config"
	"gala/lastfm"
	"gala/music"
	"gala/services"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

//go:embed icon.ico
var iconData []byte

const RegistryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const AppName = "GalaScrobbler"

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
	}

	systray.SetIcon(iconData)
	systray.SetTitle("Gala")
	systray.SetTooltip("Gala")
	mStatus, mAuth, mDiscord, mStart, mLogout, mQuit := setupTrayMenu(conf)

	updateAuthUI(conf, mAuth, mLogout)

	go music.MonitorMusic(mStatus, conf)
	go handleEvents(conf, mStatus, mDiscord)
	go handleMenuClicks(mAuth, mLogout, mQuit, mDiscord, mStart, conf)
}

func setupTrayMenu(conf *config.Config) (*systray.MenuItem, *systray.MenuItem, *systray.MenuItem, *systray.MenuItem, *systray.MenuItem, *systray.MenuItem) {
	mStatus := systray.AddMenuItem("Status: Idle", "Current status")
	mStatus.Disable()
	systray.AddSeparator()

	mAuth := systray.AddMenuItem("Login to last.fm", "Authenticate")
	mDiscord := systray.AddMenuItemCheckbox("Enable Discord RPC", "Toggle Discord", conf.GetPresenceCheck())
	mStart := systray.AddMenuItemCheckbox("Run at Startup", "Launch on login", IsAutostartEnabled())
	mLogout := systray.AddMenuItem("Logout", "Close session")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Close app")

	return mStatus, mAuth, mDiscord, mStart, mLogout, mQuit
}

func updateAuthUI(conf *config.Config, mAuth, mLogout *systray.MenuItem) {
	if conf.GetSessionKey() != "" {
		mAuth.SetTitle("Logged in as: " + conf.GetUsername())
		mAuth.Disable()
		mLogout.Enable()
	} else {
		mAuth.SetTitle("Login to last.fm")
		mAuth.Enable()
		mLogout.Disable()
	}
}

func handleEvents(conf *config.Config, mStatus *systray.MenuItem, mDiscord *systray.MenuItem) {
	discordS := services.NewDiscordService(conf.GetDiscordAppId())

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case track := <-music.TrackEventChan:
			runLogic(track, conf, mStatus, mDiscord, discordS)

		case <-ticker.C:
			current := music.GetTrack()
			if current.Title != "" {
				runLogic(current, conf, mStatus, mDiscord, discordS)
			}
		}
	}
}

func runLogic(track music.TrackInfo, conf *config.Config, mStatus *systray.MenuItem, mDiscord *systray.MenuItem, ds *services.DiscordService) {
	if track.IsPlaying {
		mStatus.SetTitle(track.Title)
	} else {
		mStatus.SetTitle("Paused")
	}

	ds.Update(track, mDiscord.Checked())

	if conf.GetSessionKey() != "" {
		services.Manager.ProcessScrobble(track, conf.GetSessionKey())
	}
}

func handleMenuClicks(mAuth, mLogout, mQuit, mDiscord, mStart *systray.MenuItem, conf *config.Config) {
	for {
		select {
		case <-mAuth.ClickedCh:
			updated, _ := config.LoadConfig()
			conf.SetApiKey(updated.GetApiKey())
			conf.SetSecret(updated.GetSecrect())

			if conf.GetApiKey() == "" {
				openConfigJSON()
			} else {
				attemptLogin(mAuth, mLogout, conf)
			}

		case <-mDiscord.ClickedCh:
			updated, _ := config.LoadConfig()
			conf.SetDiscordAppID(updated.GetDiscordAppId())

			if mDiscord.Checked() {
				mDiscord.Uncheck()
				conf.SetPresenceCheck(false)
			} else {
				if conf.GetDiscordAppId() == "" {
					openConfigJSON()
					mDiscord.Uncheck()
				} else {
					mDiscord.Check()
					conf.SetPresenceCheck(true)
				}
			}

		case <-mStart.ClickedCh:
			if mStart.Checked() {
				mStart.Uncheck()
				SetAutostart(false)
			} else {
				mStart.Check()
				SetAutostart(true)
			}

		case <-mLogout.ClickedCh:
			conf.SetSessionKey("")
			conf.SetUsername("")
			updateAuthUI(conf, mAuth, mLogout)

		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func attemptLogin(mAuth, mLogout *systray.MenuItem, conf *config.Config) {
	token, err := lastfm.StartAuthServer(conf.GetApiKey())
	if err != nil {
		return
	}

	session, user, err := lastfm.FetchSessionKey(token, conf.GetApiKey(), conf.GetSecrect())
	if err != nil {
		return
	}

	conf.SetSessionKey(session)
	conf.SetUsername(user)

	updateAuthUI(conf, mAuth, mLogout)
}

func SetAutostart(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, RegistryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enabled {
		executable, _ := os.Executable()
		return k.SetStringValue(AppName, executable)
	} else {
		return k.DeleteValue(AppName)
	}
}

func IsAutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, RegistryPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	_, _, err = k.GetStringValue(AppName)
	return err == nil
}

func onExit() {
	os.Exit(0)
}

func openConfigJSON() {
	path := config.GetConfigPath()
	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "start", path).Run()
	}
}
