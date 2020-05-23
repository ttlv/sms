package admin

import (
	"github.com/jinzhu/gorm"
	qor_admin "github.com/qor/admin"
	"github.com/ttlv/common_utils/readonly"
	"github.com/ttlv/sms/config"
)

type SendParamForm struct {
	Phone   string
	Content string
}

func New() *qor_admin.Admin {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	adm := qor_admin.New(&qor_admin.AdminConfig{
		SiteName: "SMS",
		DB:       db,
	})
	readonly.Setup(adm)
	configSmsRecoardRes(adm)
	configSmsFailureRecordRes(adm)
	configBrandRes(adm)
	configSmsSettingRes(adm)
	configSmsAvailableRes(adm)
	return adm
}
