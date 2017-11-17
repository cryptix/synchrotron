package i18n

import (
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"

	"github.com/cryptix/synchrotron/db"
)

var I18n *i18n.I18n

func init() {
	I18n = i18n.New(database.New(db.DB))
}
