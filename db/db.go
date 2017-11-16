package db

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/l10n"
	"github.com/qor/media"
	"github.com/qor/publish2"
	"github.com/qor/sorting"
	"github.com/qor/validations"

	"github.com/cryptix/synchrotron/config"
)

var (
	DB *gorm.DB
)

func init() {
	var err error

	dbConfig := config.Config.DB
	DB, err = gorm.Open("sqlite3", fmt.Sprintf("%v/%v", os.TempDir(), dbConfig.Name))
	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		DB.LogMode(true)
	}

	l10n.RegisterCallbacks(DB)
	sorting.RegisterCallbacks(DB)
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)
	publish2.RegisterCallbacks(DB)
}
