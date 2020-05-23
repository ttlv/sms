package test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ttlv/sms"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/rs/cors"
	"github.com/ttlv/common_utils/testingtool"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/service/producer_http/http_server"
)

var (
	db           *gorm.DB
	sessionStore sessions.CookieStore
	cfg          = config.MustGetConfig()
	err          error
)

func setup() {
	if db, err = gorm.Open("mysql", cfg.DB); err != nil {
		panic("无法连接mysql")
	}
	cs := cors.New(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3002"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	})

	sessionStore := sessions.NewCookieStore([]byte("GbeVMHok6yjFXTgDkwUzVMj"))

	router := http_server.New(db, sessionStore)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://localhost%v ==========\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, handler))
}

func runSql() {
	if db, err = gorm.Open("mysql", cfg.DB); err != nil {
		panic("无法连接mysql")
	}
	utils.RunSQL(db, `
		TRUNCATE TABLE sms_brands;
		TRUNCATE TABLE sms_availables;
		TRUNCATE TABLE sms_records;
`)
	utils.RunSQL(db, `
		INSERT INTO sms_brands(name, token) VALUES ('test', "123456");
		INSERT INTO sms_availables(sms_brand_id) VALUES (1)
`)
}

func TestApiSend(t *testing.T) {
	var (
		wg           = sync.WaitGroup{}
		maxGoRoutine = make(chan struct{}, 10)
	)
	go setup()
	runSql()
	time.Sleep(1 * time.Second)
	// 未上传brand
	result, _ := utils.Post("http://localhost:6004/send", map[string]string{
		"brand": "",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"未填写Brand\",\"code\":400}}", result)
	// 未上传phone
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand": "test",
		"phone": "",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"未填写手机号码\",\"code\":400}}", result)
	// 未填写发送内容
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand":   "test",
		"phone":   "8618000000000",
		"content": "",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"未填写需要发送的内容\",\"code\":400}}", result)

	// db未设置token但是http header上传了
	db.Exec("UPDATE sms_brands SET token='' WHERE name = 'test'")
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand":   "test",
		"phone":   "8618000000000",
		"content": "hello",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"未发现token,无法调用服务。\",\"code\":400}}", result)
	// db设置了token但是与http header上传的token不一样
	db.Exec("UPDATE sms_brands SET token='123456' WHERE name = 'test'")
	time.Sleep(1 * time.Second)
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand":   "test",
		"phone":   "8618000000000",
		"content": "hello",
	}, nil, map[string]string{
		"Authorization": "Token token=00000",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"未发现token,无法调用服务。\",\"code\":400}}", result)
	//短信的可用条数为0
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand":   "test",
		"phone":   "8618000000000",
		"content": "hello",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	testingtool.CompareRecords(t, "{\"error\":{\"message\":\"您无可用的短信条数,请联系管理员充值.\",\"code\":400}}", result)
	// db中可用的短信条数为1,正常发送一条之后可用条数减为0而且会在sms_records表创建一条记录
	db.Exec("UPDATE sms_availables SET available_amount = 1 WHERE sms_brand_id = 1")
	result, _ = utils.Post("http://localhost:6004/send", map[string]string{
		"brand":   "test",
		"phone":   "8618000000000",
		"content": "hello",
	}, nil, map[string]string{
		"Authorization": "Token token=123456",
	})
	var count int64
	smsAvailable := sms.SmsAvailable{}
	db.Model(&sms.SmsRecord{}).Where("phone = '+86 180 0000 0000'").Count(&count)
	db.First(&smsAvailable)
	testingtool.CompareRecords(t, "1", fmt.Sprintf("%v", count))
	testingtool.CompareRecords(t, "0", fmt.Sprintf("%v", smsAvailable.AvailableAmount))
	// 并发环境下测试短信可用余额数是否会出现异常
	// 假设一秒钟发起了10次http请求,现在的短信可用余额为10,正常的流程下执行完之后可用余额是0
	db.Exec(`UPDATE sms_availables SET available_amount = 10 WHERE sms_brand_id = 1`)
	db.Exec(`TRUNCATE TABLE sms_records`)
	for i := 0; i <= 9; i++ {
		wg.Add(1)
		maxGoRoutine <- struct{}{}
		go func(i int) {
			color.Green("正在执行第%v次携程并发模拟,开始调用服务", i+1)
			utils.Post("http://localhost:6004/send", map[string]string{
				"brand":   "test",
				"phone":   "8618000000000",
				"content": "hello",
			}, nil, map[string]string{
				"Authorization": "Token token=123456",
			})
			<-maxGoRoutine
			wg.Done()
			color.Red("第%v次携程并发模拟结束,调用服务完毕", i+1)
		}(i)
		wg.Wait()
	}
	db.Model(&sms.SmsRecord{}).Where("phone = '+86 180 0000 0000'").Count(&count)
	db.First(&smsAvailable)
	testingtool.CompareRecords(t, "10", fmt.Sprintf("%v", count))
	testingtool.CompareRecords(t, "0", fmt.Sprintf("%v", smsAvailable.AvailableAmount))
}
