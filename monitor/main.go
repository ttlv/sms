package main

import (
	"net/http"
	"net/url"

	"time"

	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/robfig/cron"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
)

func main() {
	cronJob := cron.New()
	cronJob.AddFunc(fmt.Sprintf("0 */5 * * * *"), func() {
		var count int
		var counts int
		var successCount int
		db, err := gorm.Open("mysql", config.MustGetConfig().DB)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		failureSmsRecord := []sms.SmsRecord{}
		sendingSmsRecords := []sms.SmsRecord{}
		successSmsRecords := []sms.SmsRecord{}
		now := time.Now()
		m, _ := time.ParseDuration("-5m")
		before5Min := now.Add(m)
		db.Where("state = ? AND created_at BETWEEN ? AND ?", sms.RecordState_Failure, before5Min, now).Find(&failureSmsRecord).Count(&count)
		db.Where("state = ? AND created_at BETWEEN ? AND ?", sms.RecordState_Sending, before5Min, now).Find(&sendingSmsRecords).Count(&counts)
		db.Where("state = ? AND created_at BETWEEN ? AND ?", sms.RecordState_Success, before5Min, now).Find(&successSmsRecords).Count(&successCount)
		if count > 0 {
			title := "SMS 服务告警"
			message := "当前五分钟内发现有发送失败的短信，请检查服务是否正常运行!!!"
			sound := "updown"
			PushOverMonitor(title, message, sound)
		}

		if counts > 20 {
			title := "SMS 服务告警"
			message := "当前5分钟内处于sending状态的短信条数已经累计超过20条，请检查服务运行是否正常!!!"
			sound := "echo"
			PushOverMonitor(title, message, sound)
		}

		if successCount > 20 {
			title := "SMS 服务告警"
			message := "当前5分钟内处于Success状态的短信条数已经累计超过20条，请检查WebHook服务是否正常运行！！！"
			sound := "alien"
			PushOverMonitor(title, message, sound)
		}
		//判断monitor是否活着,每一个小时的第十五分钟发送状态，表示monitor运行是否正常
		if now.Minute() > 14 && now.Minute() < 16 {
			title := "SMS 监控"
			message := "SMS服务处于正常运行状态，如果下个小时的15分钟前后没有收到此消息，请检查SMS服务是否正常"
			sound := "cashregister"
			PushOverMonitor(title, message, sound)
		}
	})
	cronJob.Start()
	http.ListenAndServe(":8082", nil)
}

func PushOverMonitor(title string, message string, sound string) {
	c := config.MustGetConfig()
	http.PostForm("https://api.pushover.net/1/messages.json", url.Values{
		"token":   []string{c.APITOKEN},
		"user":    []string{c.APIKEY},
		"title":   {title},
		"message": {message},
		"sound":   {sound},
	})
}
