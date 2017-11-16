package admin

import (
	"github.com/qor/filebox"
	"github.com/qor/roles"

	"github.com/cryptix/synchrotron/config"
	"github.com/cryptix/synchrotron/config/auth"
)

var Filebox *filebox.Filebox

func init() {
	Filebox = filebox.New(config.Root + "/public/downloads")
	Filebox.SetAuth(auth.AdminAuth{})
	dir := Filebox.AccessDir("/")
	dir.SetPermission(roles.Allow(roles.Read, "admin"))
}
