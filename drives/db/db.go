package db

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"time"
)

func Connect(dbConf map[string]interface{}) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	switch dbConf["drive"] {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=Local",
			dbConf["user"],
			dbConf["pass"],
			dbConf["host"],
			dbConf["port"],
			dbConf["dbname"],
			dbConf["charset"],
			dbConf["parsetime"])
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		break
	case "pgsql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			dbConf["host"],
			dbConf["user"],
			dbConf["pass"],
			dbConf["dbname"],
			dbConf["port"],
			dbConf["sslmode"],
			dbConf["timezone"])
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		break
	case "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			dbConf["user"],
			dbConf["pass"],
			dbConf["host"],
			dbConf["port"],
			dbConf["dbname"])
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		break
	default:
		return nil, errors.New(fmt.Sprintf("not support %s database drive", dbConf["drive"]))
	}

	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(dbConf["connpool"].(int) / 10)
	sqlDB.SetMaxOpenConns(dbConf["connpool"].(int))
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		return nil, err
	}

	return db, nil
}
