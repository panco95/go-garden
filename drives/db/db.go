package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func Connect(user, pass, addr, dbname, charset string, parseTime bool, maxConn int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%v&loc=Local", user, pass, addr, dbname, charset, parseTime)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(maxConn / 10)
	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
