package auth

import (
	"context"
	"fmt"

	"github.com/waro163/wechat-bot-sdk/api"
	"github.com/waro163/wechat-bot-sdk/common"
)

// QRAuthenticator handles QR code authentication
type QRAuthenticator struct {
	apiClient api.IQrClient
	logger    common.Logger
}

// NewQRAuthenticator creates a new QR authenticator
func NewQRAuthenticator(apiClient api.IQrClient, logger common.Logger) *QRAuthenticator {
	return &QRAuthenticator{
		apiClient: apiClient,
		logger:    logger,
	}
}

// startLogin initiates QR code login
func (a *QRAuthenticator) startLogin(ctx context.Context, accountID string) (*common.GetQRCodeResponse, error) {
	a.logger.Info("Starting QR code login", common.Field{Key: "accountID", Value: accountID})

	// Fetch QR code
	qrResp, err := a.apiClient.GetQRCode(ctx, common.DefaultBotType)
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}

	a.logger.Info("QR code received",
		common.Field{Key: "qrcodeLen", Value: len(qrResp.QRCode)},
		common.Field{Key: "imgContentLen", Value: len(qrResp.QRCodeImgContent)},
	)

	return qrResp, nil
}

func (a *QRAuthenticator) loginAndDisplayQRCode(ctx context.Context, accountID string, qrCodeDisplay func(string) error) (*common.GetQRCodeResponse, error) {
	// Start login and get QR code
	qrCodeResp, err := a.startLogin(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Display QR code to user
	if qrCodeDisplay != nil {
		if err := qrCodeDisplay(qrCodeResp.QRCodeImgContent); err != nil {
			return nil, fmt.Errorf("failed to display QR code: %w", err)
		}
	}
	return qrCodeResp, nil
}

// waitForLogin waits for QR code to be scanned and confirmed
func (a *QRAuthenticator) waitForLogin(ctx context.Context, qrcode string, accountID string, qrCodeDisplay func(string) error) (*AuthResult, error) {
	a.logger.Info("Waiting for QR code scan", common.Field{Key: "accountID", Value: accountID})

	refreshCount := 0

	for {
		// Poll QR status
		statusResp, err := a.apiClient.GetQRStatus(ctx, qrcode, common.QRLongPollTimeout)
		if err != nil {
			a.logger.Error("GetQRStatus error", common.Field{Key: "error", Value: err.Error()})
			continue
			// return nil, fmt.Errorf("failed to poll QR status: %w", err)
		}

		a.logger.Debug("QR status", common.Field{Key: "status", Value: statusResp.Status})

		switch common.QRStatus(statusResp.Status) {
		case common.QRStatusWait:
			// Continue polling
			continue

		case common.QRStatusScanned:
			a.logger.Info("QR code scanned, waiting for confirmation")
			continue

		case common.QRStatusConfirmed:
			a.logger.Info("QR code confirmed, login successful")

			if statusResp.BotToken == "" {
				return nil, fmt.Errorf("bot_token is empty in confirmed response")
			}

			return &AuthResult{
				AccountID: accountID,
				Token:     statusResp.BotToken,
				BaseURL:   statusResp.BaseURL,
				BotID:     statusResp.ILinkBotID,
				UserID:    statusResp.ILinkUserID,
			}, nil

		case common.QRStatusExpired:
			a.logger.Warn("QR code expired", common.Field{Key: "refreshCount", Value: refreshCount})

			if refreshCount >= common.MaxQRRefreshCount {
				return nil, fmt.Errorf("QR code expired after %d refresh attempts", common.MaxQRRefreshCount)
			}

			// Refresh QR code
			qrResp, err := a.loginAndDisplayQRCode(ctx, accountID, qrCodeDisplay)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh QR code: %w", err)
			}

			qrcode = qrResp.QRCode
			refreshCount++

			a.logger.Info("QR code refreshed",
				common.Field{Key: "refreshCount", Value: refreshCount},
				common.Field{Key: "newQRCode", Value: qrResp.QRCode},
				common.Field{Key: "newQRCodeImgContent", Value: qrResp.QRCodeImgContent},
			)

			continue

		default:
			return nil, fmt.Errorf("unknown QR status: %s", statusResp.Status)
		}
	}
}

// Authenticate performs complete QR code authentication flow
func (a *QRAuthenticator) Authenticate(ctx context.Context, accountID string, qrCodeDisplay func(string) error) (*AuthResult, error) {
	qrCodeResp, err := a.loginAndDisplayQRCode(ctx, accountID, qrCodeDisplay)
	if err != nil {
		return nil, err
	}
	// Wait for scan and confirmation
	result, err := a.waitForLogin(ctx, qrCodeResp.QRCode, accountID, qrCodeDisplay)
	if err != nil {
		return nil, err
	}
	return result, nil
}
