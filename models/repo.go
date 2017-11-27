package models

import "github.com/jinzhu/gorm"

type Repository struct {
	gorm.Model
	Name, URL string
	Type      string
	Heads     []BranchHead
}

type BranchHead struct {
	gorm.Model
	Name, Hash string
}
