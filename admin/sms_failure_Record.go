package admin

import (
	"github.com/jinzhu/gorm"
	qor_admin "github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"github.com/ttlv/common_utils"
	"github.com/ttlv/sms"
)

func configSmsFailureRecordRes(adm *qor_admin.Admin) {
	res := adm.AddResource(&sms.SmsFailureRecord{})
	res.UseTheme("readonly")
	common_utils.ReplaceFindManyHandler(res, "sms_failure_records")
	res.SearchAttrs("")
	searchHandler := res.SearchHandler
	res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
		context.SetDB(context.DB.Preload("SmsRecord"))
		return searchHandler(keyword, context)
	}

	res.Filter(&qor_admin.Filter{
		Name:  "Phone",
		Label: "手机号码",
		Handler: func(db *gorm.DB, arg *qor_admin.FilterArgument) *gorm.DB {
			var (
				phoneNumber = arg.Value.Get("Value").Value.([]string)[0]
			)
			_, raw, _ := parsePhoneNumber(phoneNumber)
			return db.Where("phone = ?", raw)
		},
	})

	res.Meta(&qor_admin.Meta{
		Name:  "CreatedAt",
		Label: "创建时间",
	})

	res.Meta(&qor_admin.Meta{
		Name:  "Phone",
		Label: "手机号",
	})

	res.Meta(&qor_admin.Meta{
		Name:  "ProviderName",
		Label: "短信服务商",
	})

	res.Meta(&qor_admin.Meta{
		Name:  "Error",
		Label: "错误",
	})

	res.IndexAttrs("CreatedAt", "Brand", "Phone", "ProviderName", "Error")
	res.Permission = roles.Deny(roles.Create, roles.Anyone).Deny(roles.Update, roles.Anyone).Deny(roles.Delete, roles.Anyone)
}
