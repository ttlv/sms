package server_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/theplant/testingutils"
	"github.com/ttlv/common_utils/testingtool"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"
)

type TestCase struct {
	Country       string
	ProviderSorts string
	BrandEnables  string
	ExpectedSort  string
	Content       string
}

func TestSortAndEnableProvider(t *testing.T) {
	cfg := config.MustGetConfig()
	conn, err := amqp.Dial(cfg.AMQPDial)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		panic(err)
	}
	smsQueue := amqp_queue.New(conn, ch)
	setup(smsQueue)
	defer DB.Close()

	utils.RunSQL(DB, `
	  TRUNCATE TABLE sms_brands;
	  INSERT INTO sms_brands (name, enable_emay, enable_twilio, enable_yun_pian, enable_aws) VALUES ('OCX', false, false, false, false);
	`)
	testcases := []TestCase{
		{
			ProviderSorts: "YunPian, Emay, Twilio, Amazon",
			BrandEnables:  "YunPian:true; Emay:false; Twilio:true; Amazon:true",
			ExpectedSort:  "YunPian; Twilio; Amazon",
		},
		{
			ProviderSorts: "Emay, YunPian, Amazon, Twilio",
			BrandEnables:  "YunPian:true; Emay:true; Twilio:true; Amazon:true",
			ExpectedSort:  "Emay; YunPian; Amazon; Twilio",
		},
		{
			ProviderSorts: "Emay, YunPian, Amazon, Twilio",
			BrandEnables:  "YunPian:true; Emay:false; Amazon:true; Twilio:true",
			ExpectedSort:  "YunPian; Amazon; Twilio",
		},
	}
	for i, testCase := range testcases {
		var (
			setting = &sms.SmsSetting{}
			brand   = &sms.SmsBrand{}
		)
		DB.First(&setting)
		setting.ProviderSorts = testCase.ProviderSorts
		DB.Save(setting)
		DB.Exec(`TRUNCATE TABLE sms_failure_records`)

		setProviderEnable(testCase.BrandEnables)
		time.Sleep(600 * time.Millisecond)

		DB.First(&brand)
		SmsServer.Send(context.Background(), &sms.SendParams{Phone: "8620000000000", Brand: "OCX", Country: "", Content: "C6"})
		time.Sleep(1 * time.Second) //等待写入 失败数据
		got := testingtool.GetRecords(DB, "sms_failure_records", "provider_name")
		if diff := testingutils.PrettyJsonDiff(testCase.ExpectedSort, got); len(diff) > 0 {
			t.Errorf("TestSortAndEnableProvider #%v: %v", i+1, diff)
		}
	}
}

func setProviderEnable(enables string) {
	brand := &sms.SmsBrand{}
	DB.First(&brand)
	providers := strings.Split(enables, ";")
	for _, provider := range providers {
		datas := strings.Split(strings.TrimSpace(provider), ":")
		switch strings.TrimSpace(datas[0]) {
		case "Emay":
			brand.EnableEmay = strings.TrimSpace(datas[1]) == "true"
		case "Twilio":
			brand.EnableTwilio = strings.TrimSpace(datas[1]) == "true"
		case "YunPian":
			brand.EnableYunPian = strings.TrimSpace(datas[1]) == "true"
		case "Amazon":
			brand.EnableAWS = strings.TrimSpace(datas[1]) == "true"
		}
	}
	DB.Save(brand)
}
