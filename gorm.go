package melware

import (
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
	"github.com/jinzhu/gorm"
)

type Gorm struct {
	*viper.Viper

}

func NewGorm(mel *mel.Mel) *Gorm {
	dbConfig := DBConfig(mel)
	db, err := gorm.Open(dbConfig.GetString("driver"), dbConfig.GetString("data_source"))
	if err != nil {
		panic(err)
	}

	gorm := &Gorm{
		config.Sub("gorm"),
	}

	return gorm
}
