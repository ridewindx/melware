package melware

import (
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
	"github.com/jinzhu/gorm"
	"time"
)

type Gorm struct {
	*viper.Viper
	*gorm.DB
}

func NewGorm(mel *mel.Mel) *Gorm {
	dbConfig := DBConfig(mel)
	db, err := gorm.Open(dbConfig.GetString("driver"), dbConfig.GetString("data_source"))
	if err != nil {
		panic(err)
	}

	db.DB().SetConnMaxLifetime(time.Duration(dbConfig.GetInt64("conn_max_lifetime")))
	db.DB().SetMaxIdleConns(dbConfig.GetInt("max_idle_conns"))
	db.DB().SetMaxOpenConns(dbConfig.GetInt("max_open_conns"))

	g := &Gorm{
		GetConfig(mel).Sub("gorm"),
		db,
	}

	return g
}
