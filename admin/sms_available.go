package admin

import (
	"fmt"
	"github.com/jinzhu/gorm"
	qor_admin "github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"github.com/ttlv/sms"
)

func configSmsAvailableRes(adm *qor_admin.Admin) {
	res := adm.AddResource(&sms.SmsAvailable{})
	res.UseTheme("readonly")
	res.SearchAttrs("")
	searchHandler := res.SearchHandler
	res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
		context.SetDB(context.DB.Preload("SmsBrand"))
		return searchHandler(keyword, context)
	}
	res.Meta(&qor_admin.Meta{Name: "BrandName", FieldName: "SmsBrand.Name"})
	res.Meta(&qor_admin.Meta{
		Name: "SmsBranID",
		Config: &qor_admin.SelectOneConfig{
			Collection: func(i interface{}, c *qor.Context) (result [][]string) {
				var (
					brands = []sms.SmsBrand{}
				)
				c.DB.Find(&brands)
				for _, brand := range brands {
					result = append(result, []string{fmt.Sprintf("%v", brand.ID), brand.Name})
				}
				return
			},
		},
	})
	res.Meta(&qor_admin.Meta{Name: "Note", Type: "text"})
	res.IndexAttrs("CreatedAt", "BrandName", "AvailableAmount", "Note")
	res.NewAttrs(&qor_admin.Section{
		Rows: [][]string{
			{"SmsBranID"},
			{"AvailableAmount"},
			{"Note"},
		},
	})
	res.Permission = roles.Deny(roles.Update, roles.Anyone).Deny(roles.Delete, roles.Anyone)
}
