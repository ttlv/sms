package admin

import (
	qor_admin "github.com/qor/admin"
	"github.com/ttlv/sms"
)

func configSmsSettingRes(adm *qor_admin.Admin) {
	res := adm.AddResource(&sms.SmsSetting{}, &qor_admin.Config{Singleton: true})
	res.EditAttrs("ProviderSorts")
}
