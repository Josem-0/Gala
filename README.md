#  Gala

> **A high-performance, lightweight Windows tray utility that syncs Apple Music to Last.fm and Discord Rich Presence.**

Gala runs silently in your system tray, connecting directly to the Windows native media layer (WinRT) to read what you are currently playing on Apple Music. It updates your Discord status in real-time and automatically scrobbles your history to Last.fm.

##  Features

  * **Native WinRT Integration:** Uses a custom C++ DLL to pull accurate "Now Playing" data directly from Windows.
  * **Last.fm Scrobbling:** Fully handles Last.fm authentication, updates your "Now Playing" status, and securely scrobbles tracks.
  * **Discord Rich Presence:** Displays your current song, artist, album, and elapsed time on your Discord profile.

-----

## Installation

Gala does not require a complex installer, but it does require its companion file to work correctly.

1.  Go to the [Releases page](https://github.com/Josem-0/Gala/releases/latest) and download the latest `.zip` file.
2.  Extract the folder to a permanent location on your PC (e.g., `Documents` or `AppData`).
3.  **Important:** Ensure `media.dll` is in the exact same folder as `Gala.exe`. The app will not work without it.
4.  Double-click `Gala.exe` to start.
5.  Click the icon in your system tray to log in to Last.fm and enable Discord RPC\!

-----

##  Building from Source (For Developers)

If you want to compile Gala yourself, you will need to build both the C++ media extraction library and the Go binary.

### Prerequisites

  * [Go](https://www.google.com/search?q=https://golang.org/dl/) (1.20+)
  * Build Tools for Visual Studio (Specifically the `cl` compiler for C++)
  * [Task](https://github.com/go-task/task) (Task runner)

### Build Commands

This project uses a `Taskfile.yml` to manage builds.

**To build the release version:**

```bash
task release
```

*This compiles `media.dll` and builds the Go executable with the `-H windowsgui` flag to hide the terminal.*

**To build the debug version (Includes terminal for logs):**

```bash
task debug
```

### Known Build Issues

  * **`LNK1104: cannot open file 'build\media.dll'`**: If you see this error while building, the app is likely still running in the background. Kill the process (`Stop-Process -Name scrobbler*`) and try again.

-----

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/Josem-0/Gala/blob/main/LICENSE) file for details.
