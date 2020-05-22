package admin

import (
	qor_admin "github.com/qor/admin"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/provider/amazon"
	"github.com/ttlv/sms/provider/chuanglan"
	"github.com/ttlv/sms/provider/emay"
	"github.com/ttlv/sms/provider/twilio"
	"github.com/ttlv/sms/provider/yunpian"
)

func configBrandRes(adm *qor_admin.Admin) {
	res := adm.AddResource(&sms.SmsBrand{})
	res.Action(&qor_admin.Action{
		Name:  "TestAmazon",
		Label: "测试亚马逊",
		Handler: func(actionArgument *qor_admin.ActionArgument) error {
			params := constructParam(actionArgument)
			resp, externalID, err := amazon.New().Send(*params)
			if err != nil {
				return err
			}
			saveSmsRecord(actionArgument.Context.DB, params, "Amazon", resp, externalID)
			return nil
		},
		Resource: adm.NewResource(SendParamForm{}),
		Modes:    []string{"edit"},
	})
	res.Action(&qor_admin.Action{
		Name:  "TestEmay",
		Label: "测试亿美",
		Handler: func(actionArgument *qor_admin.ActionArgument) error {
			params := constructParam(actionArgument)
			resp, externalID, err := emay.New().Send(*params)
			if err != nil {
				return err
			}
			saveSmsRecord(actionArgument.Context.DB, params, "Emay", resp, externalID)
			return nil
		},
		Resource: adm.NewResource(SendParamForm{}),
		Modes:    []string{"edit"},
	})
	res.Action(&qor_admin.Action{
		Name:  "Test twillp",
		Label: "测试twilio",
		Handler: func(actionArgument *qor_admin.ActionArgument) error {
			params := constructParam(actionArgument)
			resp, externalID, err := twilio.New().Send(*params)
			if err != nil {
				return err
			}
			saveSmsRecord(actionArgument.Context.DB, params, "Twilio", resp, externalID)
			return nil
		},
		Resource: adm.NewResource(SendParamForm{}),
		Modes:    []string{"edit"},
	})
	res.Action(&qor_admin.Action{
		Name:  "Test yunpian",
		Label: "测试云片",
		Handler: func(actionArgument *qor_admin.ActionArgument) error {
			params := constructParam(actionArgument)
			resp, externalID, err := yunpian.New().Send(*params)
			if err != nil {
				return err
			}
			saveSmsRecord(actionArgument.Context.DB, params, "YunPian", resp, externalID)
			return nil
		},
		Resource: adm.NewResource(SendParamForm{}),
		Modes:    []string{"edit"},
	})
	res.Action(&qor_admin.Action{
		Name:  "Test ChuangLan",
		Label: "测试创蓝",
		Handler: func(argument *qor_admin.ActionArgument) error {
			params := constructParam(argument)
			resp, externalID, err := chuanglan.New().Send(*params)
			if err != nil {
				return err
			}
			saveSmsRecord(argument.Context.DB, params, "ChuangLan", resp, externalID)
			return nil
		},
		Resource: adm.NewResource(SendParamForm{}),
		Modes:    []string{"edit"},
	})

	res.IndexAttrs("Name", "TwilioAccountsID", "TwilioAuthToken", "TwilioSendNumber", "YunPianAppKey", "EmayAppID", "EmayAppKey", "AWSAccessKeyID", "AWSSecretAccessKey", "AWSRegion", "ChuangLanAccount", "ChuangLanPassword")
	res.NewAttrs(&qor_admin.Section{
		Title: "Base Config",
		Rows: [][]string{
			{"Name"},
		},
	},
		&qor_admin.Section{
			Title: "Twilio Config",
			Rows: [][]string{
				{"TwilioAccountsID", "TwilioAuthToken"},
				{"TwilioSendNumber", "EnableTwilio"},
			},
		},
		&qor_admin.Section{
			Title: "YunPian Config",
			Rows: [][]string{
				{"YunPianAppKey", "YunPianHost"},
				{"EnableYunPian"},
			},
		},
		&qor_admin.Section{
			Title: "Emay Config",
			Rows: [][]string{
				{"EmayAppID", "EmayAppKey"},
				{"EnableEmay"},
			},
		},
		&qor_admin.Section{
			Title: "Aws Config",
			Rows: [][]string{
				{"AWSAccessKeyID", "AWSSecretAccessKey"},
				{"AWSRegion", "EnableAWS"},
			},
		},
		&qor_admin.Section{
			Title: "ChuangLan",
			Rows: [][]string{
				{"ChuangLanAccount", "ChuangLanPassword"},
				{"EnableChuangLan"},
			},
		})
	res.EditAttrs(res.NewAttrs())

}
