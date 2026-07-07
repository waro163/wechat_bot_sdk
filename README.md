# WeChat Bot SDK (Go)

A powerful and easy-to-use Go SDK for building WeChat bots using the WeChat iLinkAI platform. This SDK provides a clean, idiomatic Go interface for authenticating, sending messages, and receiving messages from WeChat users.

## 🌟 Features

- **QR Code Authentication** - Easy WeChat login via QR code scanning
- **Message Handling** - Receive and process text, images, videos, voice, and file messages
- **Media Support** - Upload and download images, videos, voice messages, and files
- **Persistent Sessions** - Automatic session management and storage
- **Long Polling** - Efficient message monitoring with built-in long polling
- **Configurable** - Flexible configuration with sensible defaults
- **Extensible** - Pluggable storage, cache, and logging implementations

## 📦 Installation

```bash
go get github.com/waro163/wechat-bot-sdk
```

## 🚀 Quick Start

Here's a simple example using [Gin](https://github.com/gin-gonic/gin) to serve the QR code and handle messages:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    wechatbotsdk "github.com/waro163/wechat-bot-sdk"
    "github.com/waro163/wechat-bot-sdk/common"
)

func main() {
    var qrCodeUrl string

    // Setup HTTP server to display QR code
    engine := gin.New()
    engine.GET("/qrcode", func(c *gin.Context) {
        if qrCodeUrl == "" {
            c.String(503, "QR code not ready yet, please try again later.")
            return
        }
        c.Redirect(http.StatusTemporaryRedirect, qrCodeUrl)
    })
    go engine.Run(":8080")

    // QR code callback
    qrCodeDisplay := func(qrCodeImgContent string) error {
        qrCodeUrl = qrCodeImgContent
        fmt.Printf("Please scan the QR code to login: http://localhost:8080/qrcode\n")
        return nil
    }

    // Create client
    client, err := wechatbotsdk.New(&wechatbotsdk.Config{
        AccountID: "my-wechat-bot",
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Authenticate
    ctx := context.Background()
    authRes, err := client.Authenticate(ctx, qrCodeDisplay)
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }
    log.Printf("Authenticated: %s", authRes.AccountID)

    // Start monitoring messages
    monitor, err := client.StartMonitor()
    if err != nil {
        log.Fatalf("Failed to start monitor: %v", err)
    }
    defer monitor.Stop(5 * time.Second)

    // Handle messages
    for {
        select {
        case msg := <-monitor.Messages():
            handleMessage(client, ctx, msg)
        case err := <-monitor.Errors():
            log.Printf("Monitor error: %v", err)
        }
    }
}

func handleMessage(client *wechatbotsdk.Client, ctx context.Context, msg *common.Message) {
    // Send auto-reply
    client.SendTextMessage(ctx, msg.FromUserID, "Hello! I received your message.")

    // Process message items
    for _, item := range msg.Items {
        switch *item.Type {
        case common.MessageItemTypeText:
            log.Printf("Text: %s", *item.TextItem.Text)
        case common.MessageItemTypeImage:
            log.Printf("Received image")
        case common.MessageItemTypeVideo:
            log.Printf("Received video")
        case common.MessageItemTypeVoice:
            log.Printf("Received voice")
        case common.MessageItemTypeFile:
            log.Printf("Received file: %s", *item.FileItem.FileName)
        }
    }
}
```

## 📖 API Documentation

### Client Creation

```go
client, err := wechatbotsdk.New(&wechatbotsdk.Config{
    AccountID: "unique-bot-id",  // Required: unique identifier for your bot
    // Optional configurations:
    BaseURL:              "https://ilinkai.weixin.qq.com",
    CDNBaseURL:           "https://icdn.weixin.qq.com",
    LongPollTimeout:      35 * time.Second,
    APITimeout:           15 * time.Second,
    WorkerPoolSize:       5,
    MessageChannelBuffer: 100,
    Storage:              storage.NewFileStorage("./.wechat-bot"),
    Cache:                cache.NewMemoryCache(),
})
```

### Authentication

```go
authResult, err := client.Authenticate(ctx, func(qrCodeURL string) error {
    // Display QR code to user
    fmt.Printf("Scan QR code: %s\n", qrCodeURL)
    return nil
})
```

The SDK automatically saves authentication state and will reuse it on subsequent runs, eliminating the need to scan QR codes repeatedly.

### Sending Messages

#### Text Message
```go
err := client.SendTextMessage(ctx, userID, "Hello, World!")
```

#### Image Message
```go
imageData, _ := os.ReadFile("image.jpg")
err := client.SendImageMessage(ctx, userID, imageData)
```

#### Video Message
```go
videoData, _ := os.ReadFile("video.mp4")
err := client.SendVideoMessage(ctx, userID, videoData)
```

#### File Message
```go
fileData, _ := os.ReadFile("document.pdf")
err := client.SendFileMessage(ctx, userID, "document.pdf", fileData)
```

### Receiving Messages

```go
monitor, err := client.StartMonitor()
if err != nil {
    log.Fatal(err)
}
defer monitor.Stop(5 * time.Second)

for {
    select {
    case msg := <-monitor.Messages():
        // Handle message
        for _, item := range msg.Items {
            switch *item.Type {
            case common.MessageItemTypeText:
                fmt.Println(*item.TextItem.Text)
            case common.MessageItemTypeImage:
                // Download image
                media := item.ImageItem.Media
                data, err := client.DownloadMedia(ctx,media)
            case common.MessageItemTypeVideo:
                // Handle video
            case common.MessageItemTypeVoice:
                // Handle voice
            case common.MessageItemTypeFile:
                // Handle file
            }
        }
    case err := <-monitor.Errors():
        log.Printf("Error: %v", err)
    }
}
```

### Downloading Media

```go
// From received message
media := imageItem.Media
mediaData, err := client.DownloadMedia(ctx,media)

// Save to file
os.WriteFile("downloaded.jpg", mediaData, 0644)
```

## ⚙️ Configuration

### Storage Options

By default, the SDK uses file-based storage in `./.wechat-bot`. You can implement your own storage:

```go
type CustomStorage struct{}

func (s *CustomStorage) SaveAccount(ctx context.Context, account *common.Account) error {
    // Save to database
    return nil
}

func (s *CustomStorage) LoadAccount(ctx context.Context, accountID string) (*common.Account, error) {
    // Load from database
    return &common.Account{}, nil
}

client, err := wechatbotsdk.New(&wechatbotsdk.Config{
    AccountID: "my-bot",
    Storage:   &CustomStorage{},
})
```

### Cache Options

The default memory cache stores context tokens. For distributed systems, implement your own cache:

```go
type RedisCache struct{}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // Set in Redis
    return nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    // Get from Redis
    return nil, nil
}

client, err := wechatbotsdk.New(&wechatbotsdk.Config{
    AccountID: "my-bot",
    Cache:     &RedisCache{},
})
```

### Custom Logger

```go
type CustomLogger struct{}

func (l *CustomLogger) Debug(msg string, fields ...common.Field) {}
func (l *CustomLogger) Info(msg string, fields ...common.Field) {}
func (l *CustomLogger) Warn(msg string, fields ...common.Field) {}
func (l *CustomLogger) Error(msg string, fields ...common.Field) {}

client, err := wechatbotsdk.New(&wechatbotsdk.Config{
    AccountID: "my-bot",
    Logger:    &CustomLogger{},
})
```

## 📁 Project Structure

```
wechat_bot_sdk/
├── api/            # API client and HTTP transport
├── auth/           # QR code authentication
├── cache/          # Cache implementations (memory, etc.)
├── cdn/            # Media upload/download
├── common/         # Shared types and utilities
├── crypto/         # Encryption/decryption utilities
├── messaging/      # Message sending
├── monitor/        # Long polling message monitor
├── storage/        # State persistence (file, etc.)
├── _examples/      # Example implementations
│   └── gin/        # Gin web framework example
├── client.go       # Main SDK client
├── config.go       # Configuration
└── README.md
```

## 🔗 Examples

See the [`_examples/gin`](_examples/gin) directory for a complete working example with:
- QR code authentication via web interface
- Message handling for all message types
- Media upload and download
- Auto-reply functionality

## 📝 Credits

This project was inspired by and references the implementation patterns from:
- [@tencent-weixin/openclaw-weixin](https://www.npmjs.com/package/@tencent-weixin/openclaw-weixin) - The official npm package for WeChat OpenClaw

## 📄 License

[MIT License](LICENSE)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📮 Support

For issues, questions, or contributions, please open an issue on GitHub.
