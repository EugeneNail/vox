package domain

import "strings"

type ImageExtension string

const (
	ImageExtensionGIF  ImageExtension = ".gif"
	ImageExtensionJPEG ImageExtension = ".jpg"
	ImageExtensionPNG  ImageExtension = ".png"
	ImageExtensionWEBP ImageExtension = ".webp"
)

func (extension ImageExtension) String() string {
	return string(extension)
}

func ImageExtensionFromContentType(contentType string) (ImageExtension, bool) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/gif":
		return ImageExtensionGIF, true
	case "image/jpeg":
		return ImageExtensionJPEG, true
	case "image/png":
		return ImageExtensionPNG, true
	case "image/webp":
		return ImageExtensionWEBP, true
	default:
		return "", false
	}
}
