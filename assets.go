package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("failed to generate random bytes")
	}

	randomId := base64.RawURLEncoding.EncodeToString(randomBytes)
	ext := mediaTypeToExt(mediaType)

	return fmt.Sprintf("%s%s", randomId, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getVideoLocation(key string) string {
	return fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	v4presignedHttpReq, err := presignClient.PresignGetObject(
		context.Background(),
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		},
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", fmt.Errorf("error getting presigned http request: %w", err)
	}

	return v4presignedHttpReq.URL, nil
}

func (cfg apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return database.Video{}, fmt.Errorf("unable to get video url")
	}
	if len(*video.VideoURL) == 0 {
		return database.Video{}, fmt.Errorf("unable to get video url")
	}

	videoLocation := strings.Split(*video.VideoURL, ",")
	if len(videoLocation) != 2 {
		return database.Video{}, fmt.Errorf("invalid video location")
	}
	videoBucket := videoLocation[0]
	videoKey := videoLocation[1]
	if len(videoBucket) == 0 || len(videoKey) == 0 {
		return database.Video{}, fmt.Errorf("invalid video location data")
	}

	presignedURL, err := generatePresignedURL(cfg.s3Client, videoBucket, videoKey, 15*time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("unable to sign video, %w", err)
	}

	video.VideoURL = &presignedURL
	return video, nil
}
