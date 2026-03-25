package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	wechatbotsdk "github.com/waro163/wechat-bot-sdk"
	"github.com/waro163/wechat-bot-sdk/common"
)

func main() {
	var qrCodeUrl string

	engine := gin.New()
	engine.GET("/qrcode", func(c *gin.Context) {
		if qrCodeUrl == "" {
			c.String(503, "QR code not ready yet, please try again later.")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, qrCodeUrl)
		// c.String(200, qrCodeUrl)
	})

	go engine.Run(":8080")

	qrCodeDisplay := func(qrCodeImgContent string) error {
		qrCodeUrl = qrCodeImgContent
		fmt.Printf("Please scan the QR code to login: %s\n", qrCodeUrl)
		return nil
	}

	client, err := wechatbotsdk.New(&wechatbotsdk.Config{
		AccountID: "weixin-claw-bot",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	authRes, err := client.Authenticate(ctx, qrCodeDisplay)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
		return
	}
	log.Printf("auth info: %#v", authRes)

	moni, err := client.StartMonitor()
	if err != nil {
		log.Fatalf("Failed to start monitor: %v", err)
		return
	}
	defer moni.Stop(5 * time.Second)

	handleMessage := func(input *common.Message) {
		client.SendTextMessage(ctx, input.FromUserID, "Hello! This is a response from the bot.")
		for _, msg := range input.Items {
			switch *msg.Type {
			case common.MessageItemTypeText:
				log.Printf("Received text message: %s\n", *msg.TextItem.Text)
			case common.MessageItemTypeImage:
				log.Printf("Received image message\n")
				spew.Dump(msg.ImageItem)
				// download image and send back as a test
				media := msg.ImageItem.Media
				mediaData, err := client.DownloadMedia(ctx, *media.EncryptQueryParam, *media.AESKey)
				if err != nil {
					log.Printf("Failed to download media: %v", err)
					continue
				}
				fileName := "./downloaded_image.jpg"
				f, err := os.Create(fileName)
				if err != nil {
					log.Printf("Failed to create file: %v", err)
					continue
				}
				defer f.Close()
				if _, err := f.Write(mediaData); err != nil {
					log.Printf("Failed to write media data to file: %v", err)
					continue
				}
				data, err := os.ReadFile(fileName)
				if err != nil {
					log.Printf("Failed to read image file: path:%s, error:%v", fileName, err)
					continue
				}
				if err := client.SendImageMessage(ctx, input.FromUserID, data); err != nil {
					log.Printf("Failed to send image message: path: %s, error: %v", fileName, err)
				}
			case common.MessageItemTypeVideo:
				log.Printf("Received video message\n")
				spew.Dump(msg.VideoItem)
				media := msg.VideoItem.Media
				mediaData, err := client.DownloadMedia(ctx, *media.EncryptQueryParam, *media.AESKey)
				if err != nil {
					log.Printf("Failed to download media: %v", err)
					continue
				}
				videoPath := "./downloaded_video.mp4"
				f, err := os.Create(videoPath)
				if err != nil {
					log.Printf("Failed to create file: %v", err)
					continue
				}
				defer f.Close()
				if _, err := f.Write(mediaData); err != nil {
					log.Printf("Failed to write media data to file: %v", err)
					continue
				}
				data, err := os.ReadFile(videoPath)
				if err != nil {
					log.Printf("Failed to read video file: path:%s, error:%v", videoPath, err)
					continue
				}
				if err := client.SendVideoMessage(ctx, input.FromUserID, data); err != nil {
					log.Printf("Failed to send video message: path: %s, error: %v", videoPath, err)
				}
			case common.MessageItemTypeVoice:
				log.Printf("Received voice message\n")
				spew.Dump(msg.VoiceItem)
				log.Printf("voice content: %s\n", *msg.VoiceItem.Text)
				media := msg.VoiceItem.Media
				mediaData, err := client.DownloadMedia(ctx, *media.EncryptQueryParam, *media.AESKey)
				if err != nil {
					log.Printf("Failed to download media: %v", err)
					continue
				}
				voicePath := "./downloaded_voice.wav"
				f, err := os.Create(voicePath)
				if err != nil {
					log.Printf("Failed to create file: %v", err)
					continue
				}
				defer f.Close()
				if _, err := f.Write(mediaData); err != nil {
					log.Printf("Failed to write media data to file: %v", err)
					continue
				}
				data, err := os.ReadFile(voicePath)
				if err != nil {
					log.Printf("Failed to read voice file: path:%s, error:%v", voicePath, err)
					continue
				}
				if err := client.SendFileMessage(ctx, input.FromUserID, "downloaded_voice.wav", data); err != nil {
					log.Printf("Failed to send voice message: path: %s, error: %v", voicePath, err)
				}
				// voicePath := "/Users/wangron/Downloads/1773132789629380.wav"
				// data, err := os.ReadFile(voicePath)
				// if err != nil {
				// 	log.Printf("Failed to read voice file: path:%s, error:%v", voicePath, err)
				// 	continue
				// }
				// if err := client.SendFileMessage(ctx, input.FromUserID, "1773132789629380.wav", data); err != nil {
				// 	log.Printf("Failed to send voice message: path: %s, error: %v", voicePath, err)
				// }
			case common.MessageItemTypeFile:
				log.Printf("Received file message\n")
				spew.Dump(msg.FileItem)
				media := msg.FileItem.Media
				log.Printf("file name: %s\n", *msg.FileItem.FileName)
				mediaData, err := client.DownloadMedia(ctx, *media.EncryptQueryParam, *media.AESKey)
				if err != nil {
					log.Printf("Failed to download media: %v", err)
					continue
				}
				filePath := fmt.Sprintf("./%s", *msg.FileItem.FileName)
				f, err := os.Create(filePath)
				if err != nil {
					log.Printf("Failed to create file: %v", err)
					continue
				}
				defer f.Close()
				if _, err := f.Write(mediaData); err != nil {
					log.Printf("Failed to write media data to file: %v", err)
					continue
				}
				data, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read file: path:%s, error:%v", filePath, err)
					continue
				}
				if err := client.SendFileMessage(ctx, input.FromUserID, *msg.FileItem.FileName, data); err != nil {
					log.Printf("Failed to send file message: path: %s, error: %v", filePath, err)
				}
				// filePath := "/Users/wangron/Downloads/invoice.pdf"
				// data, err := os.ReadFile(filePath)
				// if err != nil {
				// 	log.Printf("Failed to read file: path:%s, error:%v", filePath, err)
				// }
				// if err := client.SendFileMessage(ctx, input.FromUserID, "invoice.pdf", data); err != nil {
				// 	log.Printf("Failed to send file message: path: %s, error: %v", filePath, err)
				// }
			default:
				log.Printf("Received message of type %d", msg.Type)
			}
		}
	}

	func() {
		for {
			select {
			case msg := <-moni.Messages():
				fmt.Printf("Received message: %#v\n", *msg)
				handleMessage(msg)
			case err := <-moni.Errors():
				fmt.Printf("Monitor error: %v\n", err)
			}
		}
	}()
}
