package webhook

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/ttlv/common_utils/testingtool"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
)

type TestCase struct {
	Message    string
	Provider   string
	EnableName string
	Expected   string
}

func TestStateChange(t *testing.T) {
	var (
		db = initDB()
	)
	go func() {
		server := NewWebHookServer(db, initQueue())
		server.Run()
	}()
	time.Sleep(2 * time.Second)

	testCases := []TestCase{
		{
			Message:    `{"messageType":"DATA_MESSAGE","owner":"319250051612","logGroup":"sns/us-east-1/319250051612/DirectPublishToPhoneNumber","logStream":"a69e5e64-8e64-4873-952f-3bf879395050","subscriptionFilters":["LambdaStream_smscallback"],"logEvents":[{"id":"34366870072771344969344584909109459119270658496835092480","timestamp":1541063752212,"message":"{\"notification\":{\"messageId\":\"b66e7446-c4a1-500d-8d8a-201202da1837\",\"timestamp\":\"2018-11-01 09:15:39.529\"},\"delivery\":{\"phoneCarrier\":\"China Unicom\",\"mnc\":1,\"destination\":\"+8618668063897\",\"priceInUSD\":0.01531,\"smsType\":\"Promotional\",\"mcc\":460,\"providerResponse\":\"Message has been accepted by phone carrier\",\"dwellTimeMs\":300,\"dwellTimeMsUntilDeviceAck\":5752},\"status\":\"SUCCESS\"}"}]}`,
			Provider:   "aws",
			EnableName: "aws",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,,0
			  3,15379298840230010101,,0
			  4,15379299323960010109,,0
			  5,4214610508,,0
			  6,4232610508,,0
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,0
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `{"messageType":"DATA_MESSAGE","owner":"319250051612","logGroup":"sns/us-east-1/319250051612/DirectPublishToPhoneNumber","logStream":"a69e5e64-8e64-4873-952f-3bf879395050","subscriptionFilters":["LambdaStream_smscallback"],"logEvents":[{"id":"34366870072771344969344584909109459119270658496835092480","timestamp":1541063752212,"message":"{\"notification\":{\"messageId\":\"b66e7447-c4a1-500d-8d8a-201202da1837\",\"timestamp\":\"2018-11-01 09:15:39.529\"},\"delivery\":{\"phoneCarrier\":\"China Unicom\",\"mnc\":1,\"destination\":\"+8618668063897\",\"priceInUSD\":0.01531,\"smsType\":\"Promotional\",\"mcc\":460,\"providerResponse\":\"Unknown error attempting to reach phone\",\"dwellTimeMs\":300,\"dwellTimeMsUntilDeviceAck\":5752},\"status\":\"FAILURE\"}"}]}`,
			Provider:   "aws",
			EnableName: "aws",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,0
			  4,15379299323960010109,,0
			  5,4214610508,,0
			  6,4232610508,,0
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,0
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `reports=[{"smsId":"15379298840230010101","customSmsId":null,"state":"DELIVRD","desc":"成功","mobile":"15968130785","receiveTime":"2018-09-25 16:56:28","submitTime":"2018-09-25 16:56:25","extendedCode":null},{"smsId":"15379299323960010109","customSmsId":null,"state":"FAIL_MOBILE","desc":"运营商响应失败","mobile":"15968130785","receiveTime":"2018-09-26 10:44:47","submitTime":"2018-09-26 10:44:44","extendedCode":null}]`,
			Provider:   "emay",
			EnableName: "emay",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,0
			  6,4232610508,,0
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,0
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `sms_status=%257B%2522sid%2522%253A4214610508%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522DELIVRD%2522%252C%2522mobile%2522%253A%252218610274048%2522%252C%2522report_status%2522%253A%2522SUCCESS%2522%257D`,
			Provider:   "yunpian",
			EnableName: "yun_pian",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,,0
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,0
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `sms_status=%257B%2522sid%2522%253A4232610508%252C%2522uid%2522%253Anull%252C%2522user_receive_time%2522%253A%25222018-11-01%2B18%253A31%253A43%2522%252C%2522error_msg%2522%253A%2522UNKNOWNERROR%2522%252C%2522mobile%2522%253A%252218610274048%2522%252C%2522report_status%2522%253A%2522FAIL%2522%257D`,
			Provider:   "yunpian",
			EnableName: "yun_pian",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,YunPian: UNKNOWNERROR -- ,2
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,0
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `SmsSid=SM1fc5c260adf048999bce1a51c6a3565f&SmsStatus=delivered&MessageStatus=delivered&To=%2B8613314517617&MessageSid=SM1fc5c260adf048999bce1a51c6a3565f&AccountSid=ACa499142e597071b0adafb93e75851302&From=%2B14158779310&ApiVersion=2010-04-01`,
			Provider:   "twilio",
			EnableName: "twilio",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,YunPian: UNKNOWNERROR -- ,2
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,3
			  8,SM1fc5c260adf048999bce1a51c6a3565e,,0
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `SmsSid=SM1fc5c260adf048999bce1a51c6a3565e&SmsStatus=failed&MessageStatus=failed&To=%2B8613314517617&MessageSid=SM1fc5c260adf048999bce1a51c6a3565f&AccountSid=ACa499142e597071b0adafb93e75851302&From=%2B14158779310&ApiVersion=2010-04-01&ErrorMessage=failure`,
			Provider:   "twilio",
			EnableName: "twilio",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,YunPian: UNKNOWNERROR -- ,2
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,3
			  8,SM1fc5c260adf048999bce1a51c6a3565e,Twilio: failure -- ,2
              9,19091017034227251,,0
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `/chuanglan?receiver=null&pswd=null&msgid=19091017034227251&reportTime=1909101607&mobile=18626860751&status=DELIVRD&notifyTime=190910160719&statusDesc=%E7%9F%AD%E4%BF%A1%E5%8F%91%E9%80%81%E6%88%90%E5%8A%9F&length=1`,
			Provider:   "chuanglan",
			EnableName: "chuanglan",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,YunPian: UNKNOWNERROR -- ,2
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,3
			  8,SM1fc5c260adf048999bce1a51c6a3565e,Twilio: failure -- ,2
              9,19091017034227251,,3
              10,19091016071725341,,0
			`,
		},
		{
			Message:    `/chuanglan?receiver=null&pswd=null&msgid=19091016071725341&reportTime=1909101607&mobile=18626860751&status=FAILED&notifyTime=190910160719&statusDesc=failed&length=1`,
			Provider:   "chuanglan",
			EnableName: "chuanglan",
			Expected: `
			  1,b66e7446-c4a1-500d-8d8a-201202da1837,,3
			  2,b66e7447-c4a1-500d-8d8a-201202da1837,Amazon: Unknown error attempting to reach phone -- ,2
			  3,15379298840230010101,,3
			  4,15379299323960010109,Emay: 运营商响应失败 -- ,2
			  5,4214610508,,3
			  6,4232610508,YunPian: UNKNOWNERROR -- ,2
			  7,SM1fc5c260adf048999bce1a51c6a3565f,,3
			  8,SM1fc5c260adf048999bce1a51c6a3565e,Twilio: failure -- ,2
              9,19091017034227251,,3
              10,19091016071725341,ChuangLan: failed -- ,2
			`,
		},
	}

	for _, testCase := range testCases {
		db.Exec("UPDATE sms_brands SET enable_emay = false, enable_twilio = false, enable_yun_pian = false, enable_aws = false")
		db.Exec(fmt.Sprintf("UPDATE sms_brands SET enable_%v = true", testCase.EnableName))
		time.Sleep(1 * time.Second)
		if testCase.EnableName == "chuanglan" {
			http.Get("http://localhost:6004/" + testCase.Message)
		} else {
			http.Post("http://localhost:6004/"+testCase.Provider, "application/x-www-form-urlencoded", strings.NewReader(testCase.Message))
		}
		testingtool.CompareRecords(t, testCase.Expected, getSmsRecords(db))
	}
}

func initDB() *gorm.DB {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	db.LogMode(false)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&sms.SmsRecord{})
	utils.RunSQL(db, `
	  TRUNCATE TABLE sms_brands;
	  TRUNCATE TABLE sms_records;
	  INSERT INTO sms_brands (name, enable_emay, enable_twilio, enable_yun_pian, enable_aws) VALUES ('OCX', false, false, false, false);
	`)
	//创建amazon记录,通过messageId区分，一条成功，一条失败
	db.Create(&sms.SmsRecord{ID: 1, Brand: "OCX", ExternalID: "b66e7446-c4a1-500d-8d8a-201202da1837", ProviderResp: `{"code":0,"msg":"OK","result":{"count":1,"fee":0.041,"messageId":"b66e7446-c4a1-500d-8d8a-201202da1837"}}`, State: sms.RecordState_Sending})
	db.Create(&sms.SmsRecord{ID: 2, Brand: "OCX", ExternalID: "b66e7447-c4a1-500d-8d8a-201202da1837", ProviderResp: `{"code":0,"msg":"OK","result":{"count":1,"fee":0.041,"messageId":"b66e7447-c4a1-500d-8d8a-201202da1837"}}`, State: sms.RecordState_Sending})
	//Emay
	db.Create(&sms.SmsRecord{ID: 3, Brand: "OCX", ExternalID: "15379298840230010101", ProviderResp: `{"smsId":"15379298840230010101","mobile":"+8613337914338","customSmsId":null}`, State: sms.RecordState_Sending})
	db.Create(&sms.SmsRecord{ID: 4, Brand: "OCX", ExternalID: "15379299323960010109", ProviderResp: `{"smsId":"15379299323960010109","mobile":"+8613337914338","customSmsId":null}`, State: sms.RecordState_Sending})
	//yunpian
	db.Create(&sms.SmsRecord{ID: 5, Brand: "OCX", ExternalID: "4214610508", ProviderResp: `{"code":0,"msg":"OK","result":{"count":1,"fee":0.041,"sid":4214610508}}`, State: sms.RecordState_Sending})
	db.Create(&sms.SmsRecord{ID: 6, Brand: "OCX", ExternalID: "4232610508", ProviderResp: `{"code":0,"msg":"OK","result":{"count":1,"fee":0.041,"sid":4232610508}}`, State: sms.RecordState_Sending})
	//twilio
	db.Create(&sms.SmsRecord{ID: 7, Brand: "OCX", ExternalID: "SM1fc5c260adf048999bce1a51c6a3565f", ProviderResp: `{"sid": "SM1fc5c260adf048999bce1a51c6a3565f", "date_created": "Thu, 14 Jul 2016 14:18:40 +0000", ...}`, State: sms.RecordState_Sending})
	db.Create(&sms.SmsRecord{ID: 8, Brand: "OCX", ExternalID: "SM1fc5c260adf048999bce1a51c6a3565e", ProviderResp: `{"sid": "SM1fc5c260adf048999bce1a51c6a3565e", "date_created": "Thu, 14 Jul 2016 14:18:40 +0000", ...}`, State: sms.RecordState_Sending})
	// ChuangLan
	db.Create(&sms.SmsRecord{ID: 9, Brand: "OCX", ExternalID: "19091017034227251", ProviderResp: `{"code":"0","msgId":"19091017034227251","time":"20190910170342","errorMsg":""}`, State: sms.RecordState_Sending})
	db.Create(&sms.SmsRecord{ID: 10, Brand: "OCX", ExternalID: "19091016071725341", ProviderResp: `{"code":"1","msgId":"19091016071725341","time":"20190910170342","errorMsg":"运营商黑名单"}`, State: sms.RecordState_Sending})
	return db
}

func getSmsRecords(db *gorm.DB) string {
	var (
		records = []sms.SmsRecord{}
		results = []string{}
	)
	db.Find(&records)
	for _, record := range records {
		results = append(results, fmt.Sprintf("%v,%v,%v,%v", record.ID, record.ExternalID, strings.Replace(record.Error, ";", " --", -1), record.State))
	}
	return strings.Join(results, "; ")
}
