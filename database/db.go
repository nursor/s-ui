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
	"gorm.io/gorm/schema"
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
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("SUI_DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbUser := os.Getenv("SUI_DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("SUI_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}
	dbName := os.Getenv("SUI_DB_NAME")
	if dbName == "" {
		dbName = "sui"
	}

	var gormLogger logger.Interface
	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	// 配置 GORM 使用表前缀
	c := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "sui_",
			SingularTable: false,
		},
	}

	fmt.Println("Connecting to MySQL:", dbHost+":"+dbPort, "database:", dbName)
	var err error
	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)), c)
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

	// 配置 GORM 使用表前缀
	c := &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "sui_",
			SingularTable: false,
		},
	}

	fmt.Println("Using SQLite database:", dbPath)
	db, err = gorm.Open(sqlite.Open(dbPath), c)
	if err != nil {
		return err
	}
	if config.IsDebug() {
		db = db.Debug()
	}
	return nil
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

// GetTableName 获取带前缀的表名（用于 Raw SQL）
func GetTableName(modelName string) string {
	if db == nil {
		return "sui_" + modelName
	}
	tableName := db.NamingStrategy.TableName(modelName)
	return tableName
}
