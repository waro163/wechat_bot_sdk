package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/waro163/wechat-bot-sdk/common"
)

// HTTPClient abstracts HTTP operations for testing and customization
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type IQrClient interface {
	GetQRCode(ctx context.Context, botType string) (*common.GetQRCodeResponse, error)
	GetQRStatus(ctx context.Context, qrCode string, timeout time.Duration) (*common.GetQRStatusResponse, error)
}

type IUploadClient interface {
	HTTPClient
	GetUploadUrl(ctx context.Context, req *common.GetUploadUrlRequest) (*common.GetUploadUrlResponse, error)
}

type IMessageClient interface {
	SendMessage(ctx context.Context, req *common.SendMessageRequest) error
}

type IMonitorClient interface {
	GetUpdates(ctx context.Context, req *common.GetUpdatesRequest, timeout time.Duration) (*common.GetUpdatesResponse, error)
}

type IClient interface {
	IQrClient
	IUploadClient
	IMessageClient
	IMonitorClient
	SetBotToken(string)
}

type Client struct {
	baseUrl    *url.URL
	botToken   string
	httpClient HTTPClient
	logger     common.Logger
}

// ClientOptions holds options for creating an API client
type ClientOptions struct {
	BaseURL    string
	BotToken   string
	HTTPClient HTTPClient
	Logger     common.Logger
}

func NewClient(opts ClientOptions) (IClient, error) {
	baseUrl, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL:%s error: %w", opts.BaseURL, err)
	}
	return &Client{
		baseUrl:    baseUrl,
		botToken:   opts.BotToken,
		httpClient: opts.HTTPClient,
		logger:     opts.Logger,
	}, nil
}

// randomWechatUIN generates X-WECHAT-UIN header value
// Random uint32 -> decimal string -> base64
func randomWechatUIN() (string, error) {
	var n uint32
	if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
		return "", err
	}
	decimalStr := fmt.Sprintf("%d", n)
	return base64.StdEncoding.EncodeToString([]byte(decimalStr)), nil
}

// buildHeaders creates HTTP headers for API requests
func (c *Client) buildHeaders(bodyBytes []byte) (http.Header, error) {
	uin, err := randomWechatUIN()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UIN: %w", err)
	}

	headers := http.Header{
		"Content-Type":            []string{common.ContentTypeJSON},
		"AuthorizationType":       []string{common.AuthTypeILinkBotToken},
		"Content-Length":          []string{fmt.Sprintf("%d", len(bodyBytes))},
		"X-WECHAT-UIN":            []string{uin},
		"iLink-App-Id":            []string{common.ILinkAppId},
		"iLink-App-ClientVersion": []string{common.ILinkAppClientVersion},
	}

	if c.botToken != "" {
		headers.Set("Authorization", "Bearer "+c.botToken)
	}

	return headers, nil
}

// doRequest performs an HTTP POST request to the API
func (c *Client) doRequest(ctx context.Context, endpoint string, reqBody interface{}) ([]byte, error) {
	// Marshal request body
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	u := *c.baseUrl
	u.Path = endpoint
	url := u.String()

	// Build headers
	headers, err := c.buildHeaders(bodyBytes)
	if err != nil {
		return nil, err
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = headers

	c.logger.Debug("API request", common.Field{Key: "url", Value: url}, common.Field{Key: "endpoint", Value: endpoint})

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("API response", common.Field{Key: "status", Value: resp.StatusCode}, common.Field{Key: "bodyLen", Value: len(respBody)})

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) SetBotToken(token string) {
	c.botToken = token
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// GetQRCode fetches a QR code for authentication
func (c *Client) GetQRCode(ctx context.Context, botType string) (*common.GetQRCodeResponse, error) {
	// Build URL with query parameter
	u := *c.baseUrl
	u.Path = common.EndpointGetBotQRCode
	v := url.Values{}
	if botType != "" {
		v.Set("bot_type", botType)
	}
	u.RawQuery = v.Encode()
	url := u.String()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.logger.Debug("Fetching QR code", common.Field{Key: "botType", Value: botType})

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("QR code fetch failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var qrResp common.GetQRCodeResponse
	if err := json.Unmarshal(respBody, &qrResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &qrResp, nil
}

// GetQRStatus polls the QR code status
func (c *Client) GetQRStatus(ctx context.Context, qrCode string, timeout time.Duration) (*common.GetQRStatusResponse, error) {
	// Build URL with query parameter
	u := *c.baseUrl
	u.Path = common.EndpointGetQRStatus
	v := url.Values{}
	v.Set("qrcode", qrCode)
	u.RawQuery = v.Encode()
	url := u.String()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.logger.Debug("Polling QR status", common.Field{Key: "qrcode", Value: qrCode})
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// On timeout, return "wait" status (normal for long-poll)
		if ctx.Err() == context.DeadlineExceeded {
			return &common.GetQRStatusResponse{Status: "wait"}, nil
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("QR status poll failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var statusResp common.GetQRStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &statusResp, nil
}

// GetUpdates performs a long-poll getUpdates request
func (c *Client) GetUpdates(ctx context.Context, req *common.GetUpdatesRequest, timeout time.Duration) (*common.GetUpdatesResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Add base info
	if req.BaseInfo == nil {
		req.BaseInfo = &common.BaseInfo{ChannelVersion: common.ChannelVersion} // TODO: Get from version
	}

	respBody, err := c.doRequest(ctx, common.EndpointGetUpdates, req)
	if err != nil {
		// On timeout, return empty response (normal for long-poll)
		if ctx.Err() == context.DeadlineExceeded {
			return &common.GetUpdatesResponse{Ret: 0, Msgs: []common.WeixinMessage{}}, nil
		}
		return nil, err
	}

	var resp common.GetUpdatesResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getUpdates response: %w", err)
	}

	// Check for API errors
	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("API error: code=%d msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return &resp, nil
}

// SendMessage sends a message
func (c *Client) SendMessage(ctx context.Context, req *common.SendMessageRequest) error {
	// Add base info
	if req.BaseInfo == nil {
		req.BaseInfo = &common.BaseInfo{ChannelVersion: common.ChannelVersion}
	}

	respBody, err := c.doRequest(ctx, common.EndpointSendMessage, req)
	if err != nil {
		return err
	}

	var resp common.SendMessageResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal sendMessage response: %w", err)
	}

	if resp.ErrCode != 0 {
		return fmt.Errorf("API error: code=%d msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return nil
}

// GetUploadUrl gets CDN upload URL for media
func (c *Client) GetUploadUrl(ctx context.Context, req *common.GetUploadUrlRequest) (*common.GetUploadUrlResponse, error) {
	// Add base info
	if req.BaseInfo == nil {
		req.BaseInfo = &common.BaseInfo{ChannelVersion: common.ChannelVersion}
	}

	respBody, err := c.doRequest(ctx, common.EndpointGetUploadURL, req)
	if err != nil {
		return nil, err
	}

	var resp common.GetUploadUrlResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getUploadUrl response: %w", err)
	}

	if resp.ErrCode != 0 {
		return nil, fmt.Errorf("API error: code=%d msg=%s", resp.ErrCode, resp.ErrMsg)
	}

	return &resp, nil
}
