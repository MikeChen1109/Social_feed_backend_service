package models

type TokenResponse struct {
	Token string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}