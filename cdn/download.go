package cdn

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/common"
	"github.com/waro163/wechat-bot-sdk/crypto"
)

type Downloader struct {
	baseUrl    *url.URL
	httpClient api.HTTPClient
	logger     common.Logger
}

// NewDownloader creates a new CDN downloader
func NewDownloader(cdnBaseURL string, httpClient api.HTTPClient, logger common.Logger) (*Downloader, error) {
	baseUrl, err := url.Parse(cdnBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL:%s error: %w", cdnBaseURL, err)
	}
	return &Downloader{
		baseUrl:    baseUrl,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// DownloadAndDecrypt downloads and decrypts a media file from CDN
func (d *Downloader) downloadAndDecrypt(ctx context.Context, downloadURL, aesKeyBase64 string) ([]byte, error) {
	d.logger.Debug("Downloading from CDN", common.Field{Key: "url", Value: downloadURL})
	// Parse AES key
	aesKey, err := parseAESKey(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AES key: %w", err)
	}

	// Download encrypted data
	encrypted, err := d.fetchCDNBytes(ctx, downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	// Decrypt
	plaintext, err := crypto.DecryptAESECB(encrypted, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return plaintext, nil
}

// DownloadPlain downloads a plain (unencrypted) file from CDN
func (d *Downloader) DownloadPlain(ctx context.Context, encryptedQueryParam string) ([]byte, error) {
	downloadURL := buildDownloadURL(d.baseUrl, encryptedQueryParam)

	d.logger.Debug("Downloading plain from CDN", common.Field{Key: "url", Value: downloadURL})

	// Download data
	data, err := d.fetchCDNBytes(ctx, downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	return data, nil
}

func (d *Downloader) Download(ctx context.Context, media *common.CDNMedia) ([]byte, error) {
	if media == nil {
		return nil, fmt.Errorf("media is nil")
	}
	var downloadURL string
	if media.FullURL != nil {
		downloadURL = *media.FullURL
	} else {
		downloadURL = buildDownloadURL(d.baseUrl, *media.EncryptQueryParam)
	}
	return d.downloadAndDecrypt(ctx, downloadURL, *media.AESKey)
}

// fetchCDNBytes downloads raw bytes from CDN
func (d *Downloader) fetchCDNBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CDN download error %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// parseAESKey parses the base64-encoded AES key
// Two encodings are supported:
// 1. base64(raw 16 bytes) - images (from media.aes_key)
// 2. base64(hex string of 16 bytes) - file/voice/video (32 hex chars)
func parseAESKey(aesKeyBase64 string) ([]byte, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Case 1: Already 16 bytes (raw key)
	if len(decoded) == 16 {
		return decoded, nil
	}

	// Case 2: 32-char hex string
	if len(decoded) == 32 {
		// Check if it's valid hex
		hexPattern := regexp.MustCompile("^[0-9a-fA-F]{32}$")
		if hexPattern.Match(decoded) {
			// Parse hex string to bytes
			key, err := hex.DecodeString(string(decoded))
			if err != nil {
				return nil, fmt.Errorf("invalid hex string: %w", err)
			}
			return key, nil
		}
	}

	return nil, fmt.Errorf("aes_key must decode to 16 raw bytes or 32-char hex string, got %d bytes", len(decoded))
}

func buildDownloadURL(baseUrl *url.URL, encryptedQueryParam string) string {
	u := *baseUrl
	u.Path = path.Join(u.Path, "download")
	q := u.Query()
	q.Set("encrypted_query_param", encryptedQueryParam)
	u.RawQuery = q.Encode()
	return u.String()
}
