package cdn

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/common"
	"github.com/waro163/wechat-bot-sdk/crypto"
)

// Uploader handles media uploads to CDN
type Uploader struct {
	baseUrl    *url.URL
	httpClient api.IUploadClient
	logger     common.Logger
}

func NewUploader(cdnBaseURL string, httpClient api.IUploadClient, logger common.Logger) (*Uploader, error) {
	baseUrl, err := url.Parse(cdnBaseURL)
	if err != nil {
		return nil, err
	}
	return &Uploader{
		baseUrl:    baseUrl,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// uploadBufferToCDN uploads encrypted buffer to CDN
func (up *Uploader) uploadBufferToCDN(ctx context.Context, plaintext []byte, uploadParam, fileKey string, aesKey []byte) (string, error) {
	// Encrypt data
	ciphertext, err := crypto.EncryptAESECB(plaintext, aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Build CDN URL
	u := *up.baseUrl
	u.Path = path.Join(u.Path, "upload")
	q := u.Query()
	q.Set("encrypted_query_param", uploadParam)
	q.Set("filekey", fileKey)
	u.RawQuery = q.Encode()
	uploadURL := u.String()

	up.logger.Debug("Uploading to CDN",
		common.Field{Key: "ciphertextSize", Value: len(ciphertext)},
		common.Field{Key: "fileKey", Value: fileKey},
	)

	var downloadParam string
	var lastErr error

	// Retry loop
	for attempt := 1; attempt <= common.UploadMaxRetries; attempt++ {
		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(ciphertext))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", common.ContentTypeOctetStream)

		// Execute request
		resp, err := up.httpClient.Do(req)
		if err != nil {
			lastErr = err
			up.logger.Warn("CDN upload attempt failed",
				common.Field{Key: "attempt", Value: attempt},
				common.Field{Key: "error", Value: err.Error()},
			)
			continue
		}

		// Read response body for error messages
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check status code
		if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
			// Client error - don't retry
			errMsg := resp.Header.Get("x-error-message")
			if errMsg == "" {
				errMsg = string(body)
			}
			return "", fmt.Errorf("CDN upload client error %d: %s", resp.StatusCode, errMsg)
		}

		if resp.StatusCode != http.StatusOK {
			// Server error - retry
			errMsg := resp.Header.Get("x-error-message")
			if errMsg == "" {
				errMsg = fmt.Sprintf("status %d", resp.StatusCode)
			}
			lastErr = fmt.Errorf("CDN upload server error %d: %s", resp.StatusCode, errMsg)
			up.logger.Warn("CDN upload server error, retrying",
				common.Field{Key: "attempt", Value: attempt},
				common.Field{Key: "status", Value: resp.StatusCode},
				common.Field{Key: "error", Value: errMsg},
			)
			continue
		}

		// Extract download param from header
		downloadParam = resp.Header.Get("x-encrypted-param")
		if downloadParam == "" {
			lastErr = fmt.Errorf("CDN response missing x-encrypted-param header")
			up.logger.Warn("Missing x-encrypted-param, retrying", common.Field{Key: "attempt", Value: attempt})
			continue
		}

		up.logger.Debug("CDN upload success", common.Field{Key: "attempt", Value: attempt})
		return downloadParam, nil
	}

	return "", fmt.Errorf("CDN upload failed after %d attempts: %w", common.UploadMaxRetries, lastErr)
}

// UploadFile uploads a file to the CDN with encryption
func (up *Uploader) UploadFile(ctx context.Context, fileData []byte, toUserID string, mediaType common.UploadMediaType) (*common.UploadedFileInfo, error) {
	// Calculate MD5
	hash := md5.Sum(fileData)
	rawFileMD5 := hex.EncodeToString(hash[:])

	// Generate random AES key (16 bytes for AES-128)
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Generate random file key
	fileKeyBytes := make([]byte, 16)
	if _, err := rand.Read(fileKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate file key: %w", err)
	}
	fileKey := hex.EncodeToString(fileKeyBytes)

	// Calculate sizes
	rawSize := int64(len(fileData))
	fileSize := int64(crypto.AESECBPaddedSize(len(fileData)))

	up.logger.Debug("Uploading file",
		common.Field{Key: "rawSize", Value: rawSize},
		common.Field{Key: "fileSize", Value: fileSize},
		common.Field{Key: "md5", Value: rawFileMD5},
		common.Field{Key: "fileKey", Value: fileKey},
	)

	// Get upload URL from API
	uploadUrlReq := &common.GetUploadUrlRequest{
		FileKey:     fileKey,
		MediaType:   mediaType,
		ToUserID:    toUserID,
		RawSize:     rawSize,
		RawFileMD5:  rawFileMD5,
		FileSize:    fileSize,
		NoNeedThumb: true,
		AESKey:      hex.EncodeToString(aesKey),
	}

	uploadUrlResp, err := up.httpClient.GetUploadUrl(ctx, uploadUrlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	if uploadUrlResp.UploadParam == "" {
		return nil, fmt.Errorf("getUploadUrl returned no upload_param")
	}

	// Upload to CDN
	downloadParam, err := up.uploadBufferToCDN(ctx, fileData, uploadUrlResp.UploadParam, fileKey, aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to CDN: %w", err)
	}

	return &common.UploadedFileInfo{
		FileKey:                     fileKey,
		DownloadEncryptedQueryParam: downloadParam,
		AESKeyHex:                   hex.EncodeToString(aesKey),
		FileSize:                    rawSize,
		FileSizeCiphertext:          fileSize,
	}, nil
}
