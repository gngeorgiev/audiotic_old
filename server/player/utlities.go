package player

import vlc "github.com/adrg/libvlc-go"

func MediaStateToString(st vlc.MediaState) string {
	switch st {
	case vlc.MediaPlaying:
		return "playing"
	case vlc.MediaBuffering:
		return "buffering"
	case vlc.MediaEnded:
		return "ended"
	case vlc.MediaError:
		return "error"
	case vlc.MediaIdle:
		return "idle"
	case vlc.MediaOpening:
		return "openning"
	case vlc.MediaPaused:
		return "paused"
	case vlc.MediaStopped:
		return "stopped"
	default:
		return "unknown"
	}
}
