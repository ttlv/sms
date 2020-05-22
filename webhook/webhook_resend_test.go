package webhook

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/common_utils/testingtool"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"
	"github.com/ttlv/sms/server"
)

type ResendTestCase struct {
	PhoneNumber     string
	ProviderStatus  string
	WebHookResp     [][]string
	ERecords        string
	EFailureRecords string
}

func TestResendMessage(t *testing.T) {
	var (
		cfg          = config.MustGetConfig()
		db, err      = gorm.Open("mysql", cfg.DB)
		smsQueue     = initQueue()
		providers    = []sms.SmsProvider{}
		SmsServer, _ = server.New(db, smsQueue, providers)
	)
	if err != nil {
		panic(err)
	}
	common_utils.RunSQL(db, `
      TRUNCATE TABLE sms_records;
      TRUNCATE TABLE sms_failure_records;
      TRUNCATE TABLE sms_brands;
      TRUNCATE TABLE sms_settings;
	  INSERT INTO sms_brands (name, enable_aws, enable_twilio, enable_yun_pian, enable_emay) VALUES ('OCX', 0, 0, 1, 1);
      INSERT INTO sms_settings (id, content, provider_sorts) VALUES (1, 'provider', 'YunPian, Emay, Twilio, Amazon');
    `)
	go func() {
		server := NewWebHookServer(db, smsQueue)
		server.Run()
	}()
	time.Sleep(2 * time.Second)
	testCases := []ResendTestCase{
		{
			PhoneNumber:     `8619600000001`,
			ProviderStatus:  `YunPian: success`,
			WebHookResp:     [][]string{{"yunpian", "sms_status=%257B%2522sid%2522%253A4000%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522DELIVRD%2522%252C%2522mobile%2522%253A%252219610274048%2522%252C%2522report_status%2522%253A%2522SUCCESS%2522%257D"}},
			ERecords:        `1,4000,+86 196 0000 0001,YunPian,3,success,`,
			EFailureRecords: ``,
		},
		{
			PhoneNumber:    "8619600000002",
			ProviderStatus: `YunPian: success; Emay: success`,
			WebHookResp:    [][]string{{"yunpian", "sms_status=%257B%2522sid%2522%253A4001%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522UNKNOWNERROR%2522%252C%2522mobile%2522%253A%252219610274048%2522%252C%2522report_status%2522%253A%2522FAIL%2522%257D"}, {"emay", `reports=[{"smsId":"5001","customSmsId":null,"state":"DELIVRD","desc":"成功","mobile":"8619600000002","receiveTime":"2018-12-13 17:48:00","submitTime":"2018-12-13 17:31:58","extendedCode":null}]`}},
			//WebHookResp:    [][]string{{"yunpian", "success:1000"}, {"emay", `reports=[{"smsId":"154469351818100101","customSmsId":null,"state":"DELIVRD","desc":"成功","mobile":"8619600000002","receiveTime":"2018-12-13 17:48:00","submitTime":"2018-12-13 17:31:58","extendedCode":null}]`}},
			ERecords: `
							1,4000,+86 196 0000 0001,YunPian,3,success,
							2,5001,+86 196 0000 0002,Emay,3,success,
						`,
			EFailureRecords: `
							2,8619600000002,YunPian,UNKNOWNERROR
						`,
		},
		{
			PhoneNumber:    "8619600000003",
			ProviderStatus: `YunPian: success; Emay: failure`,
			WebHookResp:    [][]string{{"yunpian", "sms_status=%257B%2522sid%2522%253A4002%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522UNKNOWNERROR%2522%252C%2522mobile%2522%253A%252219610274048%2522%252C%2522report_status%2522%253A%2522FAIL%2522%257D"}},
			ERecords: `
						  1,4000,+86 196 0000 0001,YunPian,3,success,
			              2,5001,+86 196 0000 0002,Emay,3,success,
			              3,5002,+86 196 0000 0003,Emay,2,success,YunPian: UNKNOWNERROR
			              Emay: error
						`,
			EFailureRecords: `
						  2,8619600000002,YunPian,UNKNOWNERROR
			              3,8619600000003,YunPian,UNKNOWNERROR
			              3,+86 196 0000 0003,Emay,error
						`,
		},
		{
			PhoneNumber:    "8619600000004",
			ProviderStatus: `YunPian: success; Emay: success`,
			WebHookResp:    [][]string{{"yunpian", "sms_status=%257B%2522sid%2522%253A4003%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522UNKNOWNERROR%2522%252C%2522mobile%2522%253A%252219610274048%2522%252C%2522report_status%2522%253A%2522FAIL%2522%257D"}, {"emay", `reports=[{"smsId":"5003","customSmsId":null,"state":"Failed","desc":"失败","mobile":"8619600000004","receiveTime":"2018-12-13 17:48:00","submitTime":"2018-12-13 17:31:58","extendedCode":null}]`}},
			ERecords: `
			  1,4000,+86 196 0000 0001,YunPian,3,success,
              2,5001,+86 196 0000 0002,Emay,3,success,
              3,5002,+86 196 0000 0003,Emay,2,success,YunPian: UNKNOWNERROR
              Emay: error
              4,5003,+86 196 0000 0004,Emay,2,success,YunPian: UNKNOWNERROR
			  Emay: 失败
			`,
			EFailureRecords: `
				2,8619600000002,YunPian,UNKNOWNERROR
				3,8619600000003,YunPian,UNKNOWNERROR
				3,+86 196 0000 0003,Emay,error
				4,8619600000004,YunPian,UNKNOWNERROR
				4,8619600000004,Emay,失败
			`,
		},
	}
	for _, testCase := range testCases {
		/*
			if i != 0 {
				continue
			}
		*/
		SmsServer.Send(context.Background(), &sms.SendParams{Phone: testCase.PhoneNumber, Brand: "OCX"})
		for _, w := range testCase.WebHookResp {
			time.Sleep(1 * time.Second)
			http.Post("http://localhost:6004/"+w[0], "application/x-www-form-urlencoded", strings.NewReader(w[1]))
		}
		time.Sleep(2 * time.Second)
		smsRecord := testingtool.GetRecords(db, "sms_records", `id, external_id, phone, sender, state, provider_resp, error`)
		smsFailureRecord := testingtool.GetRecords(db, "sms_failure_records", `sms_record_id, phone, provider_name, error`)
		testingtool.CompareRecords(t, testCase.ERecords, smsRecord)
		testingtool.CompareRecords(t, testCase.EFailureRecords, smsFailureRecord)
	}
}

func initQueue() *amqp_queue.AMQPQueue {
	var (
		conf = config.MustGetConfig()
		conn *amqp.Connection
		ch   *amqp.Channel
		err  error
	)
	if err != nil {
		panic(err)
	}
	if conn, err = amqp.Dial(conf.AMQPDial); err != nil {
		panic(err)
	}

	if ch, err = conn.Channel(); err != nil {
		panic(err)
	}
	return amqp_queue.New(conn, ch)
}
