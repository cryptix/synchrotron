package config

import (
	"html/template"

	"github.com/cryptix/go/logging"
	"github.com/jinzhu/configor"
	"github.com/microcosm-cc/bluemonday"
	"github.com/qor/auth/providers/github"
	"github.com/qor/mailer"
	"github.com/qor/mailer/logger"
	"github.com/qor/redirect_back"
	"github.com/qor/render"
	"github.com/qor/session/manager"

	"github.com/cryptix/synchrotron/config/admin/bindatafs"
)

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

var Config = struct {
	Port int `default:"7000" env:"PORT"`
	DB   struct {
		Name    string `env:"DBName" default:"qor_example"`
		Adapter string
	}
	TWAK   string `env:"TWAPI_KEY" default:"key"`
	TWAS   string `env:"TWAPI_SECRET" default:"sec"`
	SMTP   SMTPConfig
	Github github.Config
}{}

var (
	View         *render.Render
	Mailer       *mailer.Mailer
	RedirectBack = redirect_back.New(&redirect_back.Config{
		SessionManager:  manager.SessionManager,
		IgnoredPrefixes: []string{"/auth"},
	})
	check = logging.CheckFatal
)

func init() {
	err := configor.Load(&Config, "config/database.yml", "config/smtp.yml", "config/application.yml")
	check(err)

	View = render.New(&render.Config{
		Layout:          "application",
		ViewPaths:       []string{"app/views"},
		AssetFileSystem: bindatafs.AssetFS,
	})

	htmlSanitizer := bluemonday.UGCPolicy()
	View.RegisterFuncMap("raw", func(str string) template.HTML {
		return template.HTML(htmlSanitizer.Sanitize(str))
	})

	// dialer := gomail.NewDialer(Config.SMTP.Host, Config.SMTP.Port, Config.SMTP.User, Config.SMTP.Password)
	// sender, err := dialer.Dial()

	// Mailer = mailer.New(&mailer.Config{
	// 	Sender: gomailer.New(&gomailer.Config{Sender: sender}),
	// })
	Mailer = mailer.New(&mailer.Config{
		Sender: logger.New(&logger.Config{}),
	})
}
