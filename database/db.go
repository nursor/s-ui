package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/alireza0/s-ui/config"
	"github.com/alireza0/s-ui/database/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func initUser() error {
	var count int64
	err := db.Model(&model.User{}).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		user := &model.User{
			Username: "admin",
			Password: "admin",
		}
		return db.Create(user).Error
	}
	return nil
}

func OpenDB(dbPath string) error {
	dbType := os.Getenv("SUI_DB_TYPE")
	if dbType == "mysql" {
		return OpenMySQLDB()
	}
	return OpenSQLiteDB(dbPath)
}

func OpenMySQLDB() error {
	dbHost := os.Getenv("SUI_DB_HOST")
	dbPort := os.Getenv("SUI_DB_PORT")
	dbUser := os.Getenv("SUI_DB_USER")
	dbPassword := os.Getenv("SUI_DB_PASSWORD")
	dbName := os.Getenv("SUI_DB_NAME")
	var err error
	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)), &gorm.Config{})
	if err != nil {
		return err
	}
	if config.IsDebug() {
		db = db.Debug()
	}
	return nil
}

func OpenSQLiteDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, 01740)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}
	db, err = gorm.Open(sqlite.Open(dbPath), c)

	if config.IsDebug() {
		db = db.Debug()
	}
	return err
}

func InitDB(dbPath string) error {
	err := OpenDB(dbPath)
	if err != nil {
		return err
	}

	// Default Outbounds
	if !db.Migrator().HasTable(&model.Outbound{}) {
		db.Migrator().CreateTable(&model.Outbound{})
		defaultOutbound := []model.Outbound{
			{Type: "direct", Tag: "direct", Options: json.RawMessage(`{}`)},
		}
		db.Create(&defaultOutbound)
	}

	err = db.AutoMigrate(
		&model.Setting{},
		&model.Tls{},
		&model.Inbound{},
		&model.Outbound{},
		&model.Service{},
		&model.Endpoint{},
		&model.User{},
		&model.Tokens{},
		&model.Stats{},
		&model.Client{},
		&model.Changes{},
	)
	if err != nil {
		return err
	}
	err = initUser()
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// GetDBType returns the current database type
func GetDBType() string {
	dbType := os.Getenv("SUI_DB_TYPE")
	if dbType == "mysql" {
		return "mysql"
	}
	return "sqlite"
}

// IsMySQL returns true if the current database is MySQL
func IsMySQL() bool {
	return GetDBType() == "mysql"
}
