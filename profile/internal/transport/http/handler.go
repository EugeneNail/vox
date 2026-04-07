package http

import "github.com/EugeneNail/vox/profile/internal/domain"

type Handler struct {
	profileRepository domain.ProfileRepository
}

func NewHandler(profileRepository domain.ProfileRepository) *Handler {
	return &Handler{
		profileRepository: profileRepository,
	}
}
