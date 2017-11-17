package auth

import (
	"time"

	"github.com/qor/auth"
	"github.com/qor/auth/authority"
	"github.com/qor/auth/providers/twitter"
	"github.com/qor/auth_themes/clean"

	"github.com/cryptix/synchrotron/app/models"
	"github.com/cryptix/synchrotron/config"
	"github.com/cryptix/synchrotron/db"
)

var (
	// Auth initialize Auth for Authentication
	Auth = clean.New(&auth.Config{
		DB:         db.DB,
		Render:     config.View,
		Mailer:     config.Mailer,
		UserModel:  models.User{},
		Redirector: auth.Redirector{RedirectBack: config.RedirectBack},
	})

	// Authority initialize Authority for Authorization
	Authority = authority.New(&authority.Config{
		Auth: Auth,
	})
)

func init() {
	//Auth.RegisterProvider(github.New(&config.Config.Github))
	Auth.RegisterProvider(twitter.New(&twitter.Config{
		ClientID:     config.Config.TWAK,
		ClientSecret: config.Config.TWAS,
	}))

	Authority.Register("logged_in_half_hour", authority.Rule{TimeoutSinceLastLogin: time.Minute * 30})
}
