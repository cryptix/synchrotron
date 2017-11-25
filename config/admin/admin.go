package admin

import (
	"fmt"

	"github.com/qor/action_bar"
	"github.com/qor/admin"
	"github.com/qor/help"
	"github.com/qor/i18n/exchange_actions"
	"github.com/qor/media/asset_manager"
	"github.com/qor/media/media_library"
	"github.com/qor/notification"
	"github.com/qor/notification/channels/database"
	"github.com/qor/page_builder"
	"github.com/qor/publish2"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/validations"
	"github.com/qor/widget"
	"golang.org/x/crypto/bcrypt"

	"github.com/cryptix/synchrotron/app/models"
	"github.com/cryptix/synchrotron/config/admin/bindatafs"
	"github.com/cryptix/synchrotron/config/auth"
	"github.com/cryptix/synchrotron/config/i18n"
	"github.com/cryptix/synchrotron/db"
)

var Admin *admin.Admin
var ActionBar *action_bar.ActionBar

func init() {
	Admin = admin.New(&admin.AdminConfig{
		SiteName: "synChrotron Admin",
		Auth:     auth.AdminAuth{},
		DB:       db.DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
		AssetFS:  bindatafs.AssetFS.NameSpace("admin"),
	})

	// Add Notification
	Notification := notification.New(&notification.Config{})
	Notification.RegisterChannel(database.New(&database.Config{DB: db.DB}))
	/* ex
	Notification.Action(&notification.Action{
		Name:         "Dismiss",
		MessageTypes: []string{"order_paid_cancelled", "info", "order_processed", "order_returned"},
		Visible: func(data *notification.QorNotification, context *admin.Context) bool {
			return data.ResolvedAt == nil
		},
		Handler: func(argument *notification.ActionArgument) error {
			return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", time.Now()).Error
		},
		Undo: func(argument *notification.ActionArgument) error {
			return argument.Context.GetDB().Model(argument.Message).Update("resolved_at", nil).Error
		},
	})
	*/
	Admin.NewResource(Notification)

	// Add Dashboard
	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	// Add Media Library
	Admin.AddResource(&media_library.MediaLibrary{}, &admin.Config{Menu: []string{"Site Management"}})

	// Add Asset Manager, for rich editor
	assetManager := Admin.AddResource(&asset_manager.AssetManager{}, &admin.Config{Invisible: true})

	// Add Help
	Help := Admin.NewResource(&help.QorHelpEntry{})
	Help.GetMeta("Body").Config = &admin.RichEditorConfig{AssetManager: assetManager}

	// Add User
	user := Admin.AddResource(&models.User{}, &admin.Config{Menu: []string{"User Management"}})
	user.Meta(&admin.Meta{Name: "Role", Config: &admin.SelectOneConfig{Collection: []string{"Admin", "Maintainer", "Member"}}})
	user.Meta(&admin.Meta{Name: "Password",
		Type:   "password",
		Valuer: func(interface{}, *qor.Context) interface{} { return "" },
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if newPassword := utils.ToString(metaValue.Value); newPassword != "" {
				bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil {
					context.DB.AddError(validations.NewError(user, "Password", "Can't encrpt password"))
					return
				}
				u := resource.(*models.User)
				u.Password = string(bcryptPassword)
			}
		},
	})
	user.Meta(&admin.Meta{Name: "Confirmed", Valuer: func(user interface{}, ctx *qor.Context) interface{} {
		if user.(*models.User).ID == 0 {
			return true
		}

		/* todo: load fromauthID
		ctx.GetDB()
		return user.(*models.User).Confirmed
		*/
		return false
	}})

	user.Filter(&admin.Filter{
		Name: "Role",
		Config: &admin.SelectOneConfig{
			Collection: []string{"Admin", "Maintainer", "Member"},
		},
	})

	user.IndexAttrs("ID", "Email", "Name", "Role")
	user.ShowAttrs(
		&admin.Section{
			Title: "Basic Information",
			Rows: [][]string{
				{"Name"},
				{"Email", "Password"},
				{"Avatar"},
				{"Role"},
				{"Confirmed"},
			},
		},
		&admin.Section{
			Title: "Credit Information",
			Rows: [][]string{
				{"Balance"},
			},
		},
		&admin.Section{
			Title: "Accepts",
			Rows: [][]string{
				{"AcceptPrivate", "AcceptLicense", "AcceptNews"},
			},
		},
		&admin.Section{
			Title: "Default Addresses",
			Rows: [][]string{
				{"DefaultBillingAddress"},
				{"DefaultShippingAddress"},
			},
		},
		"Addresses",
	)
	user.EditAttrs(user.ShowAttrs())

	// Blog Management
	article := Admin.AddResource(&models.Article{}, &admin.Config{Menu: []string{"Blog Management"}})
	article.IndexAttrs("ID", "VersionName", "ScheduledStartAt", "ScheduledEndAt", "Author", "Title")

	// Add Translations
	Admin.AddResource(i18n.I18n, &admin.Config{Menu: []string{"Site Management"}, Priority: 1})

	// Add Worker
	Worker := getWorker()
	exchange_actions.RegisterExchangeJobs(i18n.I18n, Worker)
	Admin.AddResource(Worker, &admin.Config{Menu: []string{"Site Management"}})

	// Add Setting
	Admin.AddResource(&models.Setting{}, &admin.Config{Name: "Shop Setting", Singleton: true})

	// Add Search Center Resources
	Admin.AddSearchResource(user)

	// Add ActionBar
	ActionBar = action_bar.New(Admin)
	ActionBar.RegisterAction(&action_bar.Action{Name: "Admin Dashboard", Link: "/admin"})

	initWidgets()

	PageBuilderWidgets := widget.New(&widget.Config{DB: db.DB})
	PageBuilderWidgets.WidgetSettingResource = Admin.NewResource(&QorWidgetSetting{}, &admin.Config{Name: "PageBuilderWidgets"})
	PageBuilderWidgets.WidgetSettingResource.NewAttrs(
		&admin.Section{
			Rows: [][]string{{"Kind"}, {"SerializableMeta"}},
		},
	)
	PageBuilderWidgets.WidgetSettingResource.AddProcessor(&resource.Processor{
		Handler: func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if widgetSetting, ok := value.(*QorWidgetSetting); ok {
				if widgetSetting.Name == "" {
					var count int
					context.GetDB().Set(admin.DisableCompositePrimaryKeyMode, "off").Model(&QorWidgetSetting{}).Count(&count)
					widgetSetting.Name = fmt.Sprintf("%v %v", utils.ToString(metaValues.Get("Kind").Value), count)
				}
			}
			return nil
		},
	})
	Admin.AddResource(PageBuilderWidgets)

	page := page_builder.New(&page_builder.Config{
		Admin:      Admin,
		PageModel:  &models.Page{},
		Containers: PageBuilderWidgets,
		// AdminConfig: &admin.Config{Name: "Campaign Pages or Builder", Menu: []string{"Sites & Campaign Pages"}, Priority: 2},
	})
	page.IndexAttrs("ID", "Title", "PublishLiveNow")

	// page := Admin.AddResource(&models.Page{})
	// page.Meta(&admin.Meta{
	// 	Name: "QorWidgetSettings",
	// 	Valuer: func(value interface{}, context *qor.Context) interface{} {
	// 		scope := context.GetDB().NewScope(value)
	// 		field, _ := scope.FieldByName("QorWidgetSettings")
	// 		context.GetDB().Model(value).Where("scope = ?", "default").Related(field.Field.Addr().Interface(), "QorWidgetSettings")
	// 		return field.Field.Interface()
	// 	},
	// 	Config: &admin.SelectManyConfig{
	// 		SelectionTemplate:  "metas/form/sortable_widgets.tmpl",
	// 		SelectMode:         "bottom_sheet",
	// 		DefaultCreating:    true,
	// 		RemoteDataResource: PageBuilderWidgets.WidgetSettingResource,
	// 	}})
	// page.Meta(&admin.Meta{Name: "QorWidgetSettingsSorter"})

	initFuncMap()
	initRouter()
}
