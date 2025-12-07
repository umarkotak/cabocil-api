package youtube_service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ( // ChannelDetails holds the information we want to return.
	ChannelDetails struct {
		Name         string
		ThumbnailURL string
		URL          string
	}

	// ---- Structs for parsing the API's JSON response ----

	// ApiResponse matches the top-level structure of the YouTube API response.
	ApiResponse struct {
		Items []ChannelItem `json:"items"`
	}

	// ChannelItem represents a single channel resource in the response.
	ChannelItem struct {
		Snippet Snippet `json:"snippet"`
	}

	// Snippet contains the main details like title and thumbnails.
	Snippet struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		CustomURL   string     `json:"customUrl"`
		PublishedAt string     `json:"publishedAt"`
		Thumbnails  Thumbnails `json:"thumbnails"`
		Localized   Localized  `json:"localized"`
		Country     string     `json:"country"`
	}

	// Thumbnails contains URLs for various thumbnail sizes.
	Thumbnails struct {
		Default Thumbnail `json:"default"`
		Medium  Thumbnail `json:"medium"`
		High    Thumbnail `json:"high"`
	}

	// Thumbnail contains the URL for a specific thumbnail size.
	Thumbnail struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}

	Localized struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
)

// GetYouTubeChannelDetails fetches details for a given channel ID using the YouTube Data API.
func GetYouTubeChannelDetails(apiKey, channelID string) (*ChannelDetails, error) {
	// 1. Construct the API request URL
	apiURL := "https://www.googleapis.com/youtube/v3/channels"
	params := url.Values{}
	params.Add("part", "snippet") // We need the 'snippet' to get title and thumbnails
	params.Add("id", channelID)
	params.Add("key", apiKey)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 2. Make the HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	// 3. Decode the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	// logrus.Infof("youtube response: %s", string(body))

	var apiResponse ApiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// 4. Check if any channel was found
	if len(apiResponse.Items) == 0 {
		return nil, fmt.Errorf("no channel found for ID: %s", channelID)
	}

	// 5. Extract the data and populate our struct
	channelData := apiResponse.Items[0].Snippet

	details := &ChannelDetails{
		Name:         channelData.Title,
		ThumbnailURL: channelData.Thumbnails.High.URL,
		URL:          fmt.Sprintf("https://www.youtube.com/channel/%s", channelID),
	}

	return details, nil
}
