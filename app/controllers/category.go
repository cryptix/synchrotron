package controllers

import (
	"net/http"

	"github.com/cryptix/synchrotron/app/models"
	"github.com/cryptix/synchrotron/config"
	"github.com/cryptix/synchrotron/config/utils"
)

func CategoryShow(w http.ResponseWriter, req *http.Request) {
	var (
		category models.Category
		products []models.Product
		tx       = utils.GetDB(req)
	)

	if tx.Where("code = ?", utils.URLParam("code", req)).First(&category).RecordNotFound() {
		http.Redirect(w, req, "/", http.StatusFound)
	}

	tx.Where(&models.Product{CategoryID: category.ID}).Preload("ColorVariations").Find(&products)

	config.View.Execute("category_show", map[string]interface{}{"CategoryName": category.Name, "Products": products}, req, w)
}
