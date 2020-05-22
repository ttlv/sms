package server_test

import (
	"context"
	"testing"

	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"

	"fmt"
	"time"

	"github.com/fatih/color"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/streadway/amqp"
	"github.com/theplant/testingutils"
)

func TestSendSms(t *testing.T) {
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
	utils.RunSQL(DB, `
	   INSERT INTO sms_settings (provider_sorts) VALUES ('YunPian, Emay, Twilio, Amazon');
	   INSERT INTO sms_brands (name, enable_aws, enable_twilio, enable_yun_pian, enable_emay) VALUES ('OCX', 1, 1, 1, 1);
	   INSERT INTO sms_brands (name, enable_aws, enable_twilio, enable_yun_pian, enable_emay) VALUES ('Mumain', 1, 1, 1, 1);
    `)
	time.Sleep(1 * time.Second)

	testCases := []struct {
		Brand                  string
		Country                string
		Phone                  string
		Content                string
		ExpectedSmsRecord      string
		ExpectedFailureRecords []sms.SmsFailureRecord
		ExpectedResp           sms.SendResp
	}{
		// 正确情况
		{
			Brand:                  "Mumain",
			Country:                "CN",
			Phone:                  "8618600000001",
			Content:                "C1",
			ExpectedResp:           sms.SendResp{Uid: "1", Error: ""},
			ExpectedSmsRecord:      `ID:1; Brand:Mumain; Phone: +86 186 0000 0001, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"Mumain","country":"CN","phone":"8618600000001","content":"C1"}; Sender: YunPian`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
		// 不带Country的情况
		{
			Brand:                  "Mumain",
			Country:                "",
			Phone:                  "8618600000002",
			Content:                "C2",
			ExpectedResp:           sms.SendResp{Uid: "2", Error: ""},
			ExpectedSmsRecord:      `ID:2; Brand:Mumain; Phone: +86 186 0000 0002, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"Mumain","phone":"8618600000002","content":"C2"}; Sender: YunPian`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
		// 测试带"+"的情况
		{
			Brand:                  "Mumain",
			Country:                "",
			Phone:                  "+8618600000003",
			Content:                "C3",
			ExpectedResp:           sms.SendResp{Uid: "3", Error: ""},
			ExpectedSmsRecord:      `ID:3; Brand:Mumain; Phone: +86 186 0000 0003, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"Mumain","phone":"+8618600000003","content":"C3"}; Sender: YunPian`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
		// 测试不同Brand
		{
			Brand:                  "OCX",
			Country:                "",
			Phone:                  "8618600000004",
			Content:                "C4",
			ExpectedResp:           sms.SendResp{Uid: "4", Error: ""},
			ExpectedSmsRecord:      `ID:4; Brand:OCX; Phone: +86 186 0000 0004, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"OCX","phone":"8618600000004","content":"C4"}; Sender: YunPian`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
		// 测试YunPian,,Emay,Twilio, Amazon都发不出
		{
			Brand:             "Mumain",
			Country:           "",
			Phone:             "8618600000005",
			Content:           "C5",
			ExpectedResp:      sms.SendResp{Uid: "5", Error: ""},
			ExpectedSmsRecord: `ID:5; Brand:Mumain; Phone: +86 186 0000 0005, ProviderResp: ; Error: YunPian: 发送失败;Emay: 发送失败;Twilio: 发送失败;Amazon: 发送失败;; State: 2; RawParam:{"brand":"Mumain","phone":"8618600000005","content":"C5"}; Sender: Amazon`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{
				sms.SmsFailureRecord{SmsRecordId: 5, ProviderName: "YunPian", Phone: "+86 186 0000 0005", Error: "发送失败"},
				sms.SmsFailureRecord{SmsRecordId: 5, ProviderName: "Emay", Phone: "+86 186 0000 0005", Error: "发送失败"},
				sms.SmsFailureRecord{SmsRecordId: 5, ProviderName: "Twilio", Phone: "+86 186 0000 0005", Error: "发送失败"},
				sms.SmsFailureRecord{SmsRecordId: 5, ProviderName: "Amazon", Phone: "+86 186 0000 0005", Error: "发送失败"},
			},
		},
		// 测试YunPian发不出去, Emay发出去了
		{
			Brand:             "Mumain",
			Country:           "",
			Phone:             "8618600000006",
			Content:           "C6",
			ExpectedResp:      sms.SendResp{Uid: "6", Error: ""},
			ExpectedSmsRecord: `ID:6; Brand:Mumain; Phone: +86 186 0000 0006, ProviderResp: success; Error: YunPian: 发送失败;; State: 1; RawParam:{"brand":"Mumain","phone":"8618600000006","content":"C6"}; Sender: Emay`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{
				sms.SmsFailureRecord{SmsRecordId: 6, ProviderName: "YunPian", Phone: "+86 186 0000 0006", Error: "发送失败"},
			},
		},
		// 测试YunPian, Emay都发不出去(优先使用), 用外国服务Twilio发送成功
		{
			Brand:             "Mumain",
			Country:           "",
			Phone:             "8618600000007",
			Content:           "C7",
			ExpectedResp:      sms.SendResp{Uid: "7", Error: ""},
			ExpectedSmsRecord: `ID:7; Brand:Mumain; Phone: +86 186 0000 0007, ProviderResp: success; Error: YunPian: 发送失败;Emay: 发送失败;; State: 1; RawParam:{"brand":"Mumain","phone":"8618600000007","content":"C7"}; Sender: Twilio`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{
				sms.SmsFailureRecord{SmsRecordId: 7, ProviderName: "YunPian", Phone: "+86 186 0000 0007", Error: "发送失败"},
				sms.SmsFailureRecord{SmsRecordId: 7, ProviderName: "Emay", Phone: "+86 186 0000 0007", Error: "发送失败"},
			},
		},
		// 外国号码优先使用Twillio
		{
			Brand:                  "Mumain",
			Country:                "",
			Phone:                  "+60 10-514 6182",
			Content:                "C8",
			ExpectedResp:           sms.SendResp{Uid: "8", Error: ""},
			ExpectedSmsRecord:      `ID:8; Brand:Mumain; Phone: +60 10-514 6182, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"Mumain","phone":"+60 10-514 6182","content":"C8"}; Sender: Twilio`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
		// 场景模拟 使用YunPian发送，API请求成功，等到callback，用户没有真实的收到短信，用户再次请求发送一条短信,使用第二个provider Emay优先发送
		{
			Brand:                  "Mumain",
			Country:                "",
			Phone:                  "8618600000001",
			Content:                "C9",
			ExpectedResp:           sms.SendResp{Uid: "9", Error: ""},
			ExpectedSmsRecord:      `ID:9; Brand:Mumain; Phone: +86 186 0000 0001, ProviderResp: success; Error: ; State: 1; RawParam:{"brand":"Mumain","phone":"8618600000001","content":"C9"}; Sender: Emay`,
			ExpectedFailureRecords: []sms.SmsFailureRecord{},
		},
	}
	//定义一个空数组接收db查询的结果
	SmsFailureRecords := []sms.SmsFailureRecord{}
	for i, testCase := range testCases {
		/*
			if i > 7 {
				continue
			}
		*/
		hasError := false
		resp, _ := SmsServer.Send(context.Background(), &sms.SendParams{Phone: testCase.Phone, Brand: testCase.Brand, Country: testCase.Country, Content: testCase.Content})
		if diff := testingutils.PrettyJsonDiff(testCase.ExpectedResp, resp); len(diff) > 0 {
			hasError = true
			t.Errorf(diff)
		}
		//Check ExpectedSmsRecord
		time.Sleep(1 * time.Second)
		SmsRecord := sms.SmsRecord{}
		DB.Where("id = ?", resp.Uid).Find(&SmsRecord)
		raw := fmt.Sprintf("ID:%v; Brand:%v; Phone: %v, ProviderResp: %v; Error: %v; State: %v; RawParam:%v; Sender: %v", SmsRecord.ID, SmsRecord.Brand, SmsRecord.Phone, SmsRecord.ProviderResp, SmsRecord.Error, SmsRecord.State, SmsRecord.RawParam, SmsRecord.Sender)
		if diff := testingutils.PrettyJsonDiff(testCase.ExpectedSmsRecord, raw); len(diff) > 0 {
			hasError = true
			t.Errorf(diff)
		}
		//Check ExpectedFailureRecords
		DB.Where("sms_record_id = ?", resp.Uid).Find(&SmsFailureRecords)
		if diff := testingutils.PrettyJsonDiff(testCase.ExpectedFailureRecords, SmsFailureRecords); len(diff) > 0 {
			hasError = true
			t.Errorf(diff)
		}
		if !hasError {
			fmt.Printf(color.GreenString("TestSendSms #%v: Success\n", i+1))
		}
	}
}
