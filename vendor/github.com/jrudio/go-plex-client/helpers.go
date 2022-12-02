package plex

import "errors"

// GetMediaTypeID returns plex's media type id
func GetMediaTypeID(mediaType string) string {
	switch mediaType {
	case "movie":
		return "1"
	case "show":
		return "2"
	case "season":
		return "3"
	case "episode":
		return "4"
	case "trailer":
		return "5"
	case "comic":
		return "6"
	case "person":
		return "7"
	case "artist":
		return "8"
	case "album":
		return "9"
	case "track":
		return "10"
	case "photoAlbum":
		return "11"
	case "picture":
		return "12"
	case "photo":
		return "13"
	case "clip":
		return "14"
	case "playlistItem":
		return "15"
	default:
		return mediaType
	}
}

// GetMediaType is a helper function that returns the media type. Usually, used after GetMetadata().
func GetMediaType(info MediaMetadata) string {
	if dType := info.MediaContainer.Metadata[0].Type; dType != "" {
		return dType
	}

	if vType := info.MediaContainer.Metadata[0].Type; vType != "" {
		return vType
	}

	return ""
}

// LibraryParamsFromMediaType is a helper for CreateLibraryParams
func LibraryParamsFromMediaType(mediaType string) (CreateLibraryParams, error) {
	var params CreateLibraryParams

	params.LibraryType = mediaType

	switch mediaType {
	case "movie":
		params.Agent = "com.plexapp.agents.imdb"
		params.Scanner = "Plex Movie Scanner"

		return params, nil
	case "show":
		params.Agent = "com.plexapp.agents.thetvdb"
		params.Scanner = "Plex Series Scanner"

		return params, nil
	case "music":
		params.Agent = "com.plexapp.agents.lastfm"
		params.Scanner = "Plex Music Scanner"

		return params, nil
	case "photo":
		params.Agent = "com.plexapp.agents.none"
		params.Scanner = "Plex Photo Scanner"

		return params, nil
	case "homevideo":
		params.Agent = "com.plexapp.agents.none"
		params.Scanner = "Plex Video Files Scanner"

		return params, nil
	default:
		return params, errors.New("unknown library type")
	}
}
