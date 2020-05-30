package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

//TODO as in database
type TODO struct {
	gorm.Model
	Email string `gorm:"size:60; not null"`
	Title string `gorm:"size:255; not null"`
}

func initDB() (*gorm.DB, error) {
	connectString := opt.sqliteFile

	logrus.Infof("Initializing sqlite database")
	db0, err := gorm.Open("sqlite3", connectString)
	if err != nil {
		logrus.Errorf("Couldn't connect to database. err=%s", err)
		return db0, err
	}

	if opt.logLevel == "debug" {
		db0.LogMode(true)
	}

	db0.Set("gorm:table_options", "charset=utf8")

	logrus.Infof("Checking database schema")
	db0.AutoMigrate(&TODO{})

	return db0, nil
}
