package common

// UploadMediaType represents the type of media being uploaded
type UploadMediaType int

const (
	UploadMediaTypeImage UploadMediaType = 1
	UploadMediaTypeVideo UploadMediaType = 2
	UploadMediaTypeFile  UploadMediaType = 3
	UploadMediaTypeVoice UploadMediaType = 4
)

// UploadedFileInfo represents the result of a successful file upload
type UploadedFileInfo struct {
	FileKey                     string
	DownloadEncryptedQueryParam string // CDN x-encrypted-param header
	AESKeyHex                   string // AES key in hex format
	FileSize                    int64  // Plaintext file size
	FileSizeCiphertext          int64  // Ciphertext file size (after AES-128-ECB)
}

// GetUploadUrlRequest represents the getUploadUrl API request
type GetUploadUrlRequest struct {
	FileKey         string          `json:"filekey,omitempty"`
	MediaType       UploadMediaType `json:"media_type,omitempty"`
	ToUserID        string          `json:"to_user_id,omitempty"`
	RawSize         int64           `json:"rawsize,omitempty"`
	RawFileMD5      string          `json:"rawfilemd5,omitempty"`
	FileSize        int64           `json:"filesize,omitempty"`
	ThumbRawSize    int64           `json:"thumb_rawsize,omitempty"`
	ThumbRawFileMD5 string          `json:"thumb_rawfilemd5,omitempty"`
	ThumbFileSize   int64           `json:"thumb_filesize,omitempty"`
	NoNeedThumb     bool            `json:"no_need_thumb,omitempty"`
	AESKey          string          `json:"aeskey,omitempty"`
	BaseInfo        *BaseInfo       `json:"base_info,omitempty"`
}

// GetUploadUrlResponse represents the getUploadUrl API response
type GetUploadUrlResponse struct {
	Ret              int    `json:"ret"`
	ErrCode          int    `json:"errcode,omitempty"`
	ErrMsg           string `json:"errmsg,omitempty"`
	UploadParam      string `json:"upload_param,omitempty"`
	ThumbUploadParam string `json:"thumb_upload_param,omitempty"`
}
