package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/l10n"
	"github.com/qor/location"
)

type Setting struct {
	gorm.Model
	location.Location `location:"name:Company Address"`
	l10n.Locale
}
