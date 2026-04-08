#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Media.Control.h>
#include <string>
#include <cstring>

using namespace winrt;
using namespace Windows::Media::Control;

// Cache these statically so they only initialize ONCE
static bool isInitialized = false;
static GlobalSystemMediaTransportControlsSessionManager manager = nullptr;

extern "C" __declspec(dllexport) void GetCurrentMedia(
    char* titleBuf, int titleLen, 
    char* artistBuf, int artistLen,
    char* albumBuf, int albumLen,
    char* appBuf, int appLen,
    int64_t* positionTicks, int64_t* durationTicks,
    int* playbackStatus) {

    if (titleLen > 0) titleBuf[0] = '\0';
    if (artistLen > 0) artistBuf[0] = '\0';
    if (albumLen > 0) albumBuf[0] = '\0';
    if (appLen > 0) appBuf[0] = '\0';
    *positionTicks = 0;
    *durationTicks = 0;
    *playbackStatus = 0;

    try {
        if (!isInitialized) {
            init_apartment();
            manager = GlobalSystemMediaTransportControlsSessionManager::RequestAsync().get();
            isInitialized = true;
        }

        if (!manager) return;

        auto session = manager.GetCurrentSession();
        
        if (session) {
            auto props = session.TryGetMediaPropertiesAsync().get();
            auto timeline = session.GetTimelineProperties();
            auto playback = session.GetPlaybackInfo();
            
            std::string title = to_string(props.Title());
            std::string artist = to_string(props.Artist());
            std::string album = to_string(props.AlbumTitle());
            std::string app = to_string(session.SourceAppUserModelId());

            strncpy(titleBuf, title.c_str(), titleLen - 1); titleBuf[titleLen - 1] = '\0';
            strncpy(artistBuf, artist.c_str(), artistLen - 1); artistBuf[artistLen - 1] = '\0';
            strncpy(albumBuf, album.c_str(), albumLen - 1); albumBuf[albumLen - 1] = '\0';
            strncpy(appBuf, app.c_str(), appLen - 1); appBuf[appLen - 1] = '\0';

            *positionTicks = timeline.Position().count();
            *durationTicks = timeline.EndTime().count();
            *playbackStatus = (int)playback.PlaybackStatus();
        }
    } catch (...) {
    }
}