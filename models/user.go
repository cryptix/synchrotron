package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/media"
	"github.com/qor/media/oss"
)

type User struct {
	gorm.Model
	Email    string `form:"email"`
	Password string
	Name     string `form:"name"`
	Role     string
	Avatar   AvatarImageStorage
}

func (user User) DisplayName() string        { return user.Email }
func (user User) AvailableLocales() []string { return []string{"de-DE", "en-US", "zh-CN"} }

type AvatarImageStorage struct{ oss.OSS }

func (AvatarImageStorage) GetSizes() map[string]*media.Size {
	return map[string]*media.Size{
		"small":  {Width: 50, Height: 50},
		"middle": {Width: 120, Height: 120},
		"big":    {Width: 320, Height: 320},
	}
}
