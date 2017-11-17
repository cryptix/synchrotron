package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/qor/i18n/inline_edit"
	"github.com/qor/middlewares"
	"github.com/qor/render"
	"github.com/qor/session"
	"github.com/qor/session/manager"
	"github.com/qor/widget"

	"github.com/cryptix/go/logging"
	"github.com/cryptix/synchrotron/app/models"
	"github.com/cryptix/synchrotron/config"
	"github.com/cryptix/synchrotron/config/admin"
	"github.com/cryptix/synchrotron/config/i18n"
	"github.com/cryptix/synchrotron/config/routes"
	"github.com/cryptix/synchrotron/config/utils"
	_ "github.com/cryptix/synchrotron/db/migrations"
)

var (
	log   logging.Interface
	check = logging.CheckFatal

	Revision string = "undefined"
)

func main() {
	// create timestamped logfile
	os.Mkdir("logs", 0700)
	logFileName := fmt.Sprintf("logs/%s-%s.log",
		filepath.Base(os.Args[0]),
		time.Now().Format("2006-01-02_15-04"))
	logFile, err := os.Create(logFileName)
	if err != nil {
		panic(err) // logging not ready yet...
	}
	logging.SetupLogging(io.MultiWriter(os.Stderr, logFile))
	log = logging.Logger("synchroserv")

	mux := http.NewServeMux()
	mux.Handle("/", routes.Router(kitlog.With(log, "unit", "http")))
	admin.Admin.MountTo("/admin", mux)
	admin.Filebox.MountTo("/downloads", mux)

	config.View.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}

		// Add `t` method
		for key, fc := range inline_edit.FuncMap(i18n.I18n, utils.GetCurrentLocale(req), utils.GetEditMode(w, req)) {
			funcMap[key] = fc
		}

		for key, value := range admin.ActionBar.FuncMap(w, req) {
			funcMap[key] = value
		}

		widgetContext := admin.Widgets.NewContext(&widget.Context{
			DB:         utils.GetDB(req),
			Options:    map[string]interface{}{"Request": req},
			InlineEdit: utils.GetEditMode(w, req),
		})
		for key, fc := range widgetContext.FuncMap() {
			funcMap[key] = fc
		}

		funcMap["flashes"] = func() []session.Message {
			return manager.SessionManager.Flashes(w, req)
		}

		// Add `action_bar` method
		funcMap["render_action_bar"] = func() template.HTML {
			return admin.ActionBar.Render(w, req)
		}

		funcMap["current_locale"] = func() string {
			return utils.GetCurrentLocale(req)
		}

		funcMap["current_user"] = func() *models.User {
			return utils.GetCurrentUser(req)
		}

		return funcMap
	}

	h := logging.InjectHandler(kitlog.With(log, "unit", "http"))(middlewares.Apply(mux))

	addr := fmt.Sprintf(":%d", config.Config.Port)
	log.Log("event", "listening", "addr", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		panic(err)
	}
}
