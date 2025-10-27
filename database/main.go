package database

import (
	"strings"

	"github.com/sunvc/NoLets/common"
	"gorm.io/gorm"
)

var DB Database

// Database defines all the db operation
type Database interface {
	CountAll() (int, error)                      //Get db records count
	DeviceTokenByKey(key string) (string, error) //Get specified device's token
	DeviceTokenByGroup(name string) ([]string, error)
	SaveDeviceTokenByKey(key, token, group string) (string, error) //Create or update specified device's token
	KeyExists(key string) bool
	Close() error //Close the database
}

type User struct {
	gorm.Model
	Key   string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Token string `gorm:"type:varchar(50);"`
	Group string `gorm:"type:varchar(50);column:user_group;" json:"group"`
}

func InitDatabase() {
	dsn := common.LocalConfig.System.DSN

	if len(dsn) > 5 {
		isSqlite3 = strings.Contains(dsn, "sqlite")
		if isSqlite3 {
			DB = NewSqlite3()
		} else {
			DB = NewMysql(dsn)
		}
		return
	}
	DB = NewBboltdb(common.BaseDir())
}
