package animenosub

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/google/uuid"
)

type detailsResponse struct {
	EmbedFrameUrl string `json:"embed_frame_url"`
}

type fingerprint struct {
	ViewerId   string  `json:"viewer_id"`
	DeviceId   string  `json:"device_id"`
	Confidence float32 `json:"confidence"`
	Now        int64   `json:"iat"`
	Exp        int64   `json:"exp"`
}

type fingerprintRequest struct {
	Fingerprint fingerprintData `json:"fingerprint"`
}

type fingerprintData struct {
	Token      string  `json:"token"`
	ViewerId   string  `json:"viewer_id"`
	DeviceId   string  `json:"device_id"`
	Confidence float64 `json:"confidence"`
}

type PlaybackResponse struct {
	Playback struct {
		Algorithm   string    `json:"algorithm"`
		Iv          string    `json:"iv"`
		Payload     string    `json:"payload"`
		KeyParts    []string  `json:"key_parts"`
		ExpiresAt   time.Time `json:"expires_at"`
		DecryptKeys struct {
			Edge1          string `json:"edge_1"`
			Edge2          string `json:"edge_2"`
			LegacyFallback string `json:"legacy_fallback"`
		} `json:"decrypt_keys"`
		Iv2      string `json:"iv2"`
		Payload2 string `json:"payload2"`
	} `json:"playback"`
}

type SourceData struct {
	Sources []struct {
		Quality     string `json:"quality"`
		Label       string `json:"label"`
		MimeType    string `json:"mime_type"`
		URL         string `json:"url"`
		BitrateKbps int    `json:"bitrate_kbps"`
		Height      int    `json:"height"`
		SizeBytes   int    `json:"size_bytes"`
	} `json:"sources"`
	Tracks    []interface{} `json:"tracks"`
	PosterURL string        `json:"poster_url"`
}

func videosFromUrl(videoUrl string) (string, error) {
	parsedUrl, err := url.Parse(videoUrl)
	if err != nil {
		return "", err
	}
	videoId := path.Base(parsedUrl.Path)
	detailsUrl := fmt.Sprintf("https://%s/api/videos/%s/embed/details", parsedUrl.Host, videoId)
	req, err := http.NewRequest("GET", detailsUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", videoUrl+"/")
	req.Header.Set("Origin", videoUrl)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var detailsData detailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailsData); err != nil {
		return "", err
	}

	return detailsData.EmbedFrameUrl, nil
}

func getEmbedData(embedUrl, siteUrl string) (core.StreamData, error) {
	parsedUrl, err := url.Parse(embedUrl)
	if err != nil {
		return core.StreamData{}, err
	}
	videoId := path.Base(parsedUrl.Path)

	fingerprintBody, err := buildFingerprintBody()
	if err != nil {
		return core.StreamData{}, err
	}

	bodyBytes, err := json.Marshal(fingerprintBody)
	if err != nil {
		return core.StreamData{}, err
	}

	playbackUrl := fmt.Sprintf("https://%s/api/videos/%s/embed/playback", parsedUrl.Host, videoId)
	playbackReq, err := http.NewRequest("POST", playbackUrl, bytes.NewReader(bodyBytes))
	if err != nil {
		return core.StreamData{}, err
	}

	originHost := strings.TrimPrefix(siteUrl, "https://")
	originHost = strings.TrimPrefix(originHost, "http://")
	originHost = strings.TrimSuffix(originHost, "/")
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

	playbackReq.Header.Set("Referer", embedUrl)
	playbackReq.Header.Set("Origin", "https://"+parsedUrl.Host)
	playbackReq.Header.Set("User-Agent", userAgent)

	playbackReq.Header.Set("X-Embed-Origin", originHost)
	playbackReq.Header.Set("X-Embed-Parent", fmt.Sprintf("https://%s/e/%s", parsedUrl.Host, videoId))
	playbackReq.Header.Set("X-Embed-Referer", siteUrl+"/")
	playbackReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	playbackResp, err := client.Do(playbackReq)
	if err != nil {
		return core.StreamData{}, err
	}
	defer playbackResp.Body.Close()

	var responseData PlaybackResponse
	if err := json.NewDecoder(playbackResp.Body).Decode(&responseData); err != nil {
		return core.StreamData{}, err
	}
	streamByte, err := decryptPayload(responseData)
	if err != nil {
		return core.StreamData{}, err
	}
	var source SourceData
	if err := json.Unmarshal(streamByte, &source); err != nil {
		return core.StreamData{}, err
	}
	if source.Sources[0].URL == "" {
		return core.StreamData{}, fmt.Errorf("Error: No masterUrl found")
	}
	masterUrl := source.Sources[0].URL
	header := map[string]string{
		"Referer": fmt.Sprintf("https://%s/", parsedUrl.Host),
		"Origin":  fmt.Sprintf("https://%s", parsedUrl.Host),
		// "User-Agent": userAgent,
	}
	// header.Set("User-Agent", userAgent)

	streamData := core.StreamData{
		Url:     masterUrl,
		Headers: header,
	}

	return streamData, nil
}

func decryptPayload(pb PlaybackResponse) ([]byte, error) {
	keyPart1, err := base64.RawURLEncoding.DecodeString(pb.Playback.KeyParts[0])
	if err != nil {
		return nil, err
	}

	keyPart2, err := base64.RawURLEncoding.DecodeString(pb.Playback.KeyParts[1])
	if err != nil {
		return nil, err
	}

	keyBytes := append(keyPart1, keyPart2...)

	ivBytes, err := base64.RawURLEncoding.DecodeString(pb.Playback.Iv)
	if err != nil {
		return nil, err
	}

	cipherBytes, err := base64.RawURLEncoding.DecodeString(pb.Playback.Payload)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	sourceByte, err := gcm.Open(nil, ivBytes, cipherBytes, nil)
	if err != nil {
		return nil, err
	}

	return sourceByte, nil
}

func buildFingerprintBody() (fingerprintRequest, error) {
	viewerUUID, _ := uuid.NewRandom()
	deviceUUID, _ := uuid.NewRandom()

	viewerId := strings.ReplaceAll(viewerUUID.String(), "-", "")
	deviceId := strings.ReplaceAll(deviceUUID.String(), "-", "")

	nowSec := time.Now().Unix()
	expSec := nowSec + 600

	fingerprintPayload := fingerprint{
		ViewerId:   viewerId,
		DeviceId:   deviceId,
		Confidence: 0.93,
		Now:        nowSec,
		Exp:        expSec,
	}

	rawByte, err := json.Marshal(fingerprintPayload)
	if err != nil {
		return fingerprintRequest{}, err
	}
	encodePayload := base64.RawURLEncoding.EncodeToString(rawByte)

	fingerprintToken := encodePayload + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	fingerprintBody := fingerprintRequest{
		Fingerprint: fingerprintData{
			Token:      fingerprintToken,
			ViewerId:   viewerId,
			DeviceId:   deviceId,
			Confidence: 0.93,
		},
	}
	return fingerprintBody, nil
}
