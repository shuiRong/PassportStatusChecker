package main

import (
	"flag"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jordan-wright/email"
)

func sendMail(content string) {
	println("准备发送邮件...")
	from := os.Getenv("FROM")
	to := os.Getenv("TO")
	passcode := os.Getenv("PASSWORD_CODE")
	smtpAddress := os.Getenv("SMTP")
	if smtpAddress == "" {
		smtpAddress = "smtp.qq.com"
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "25"
	}

	e := email.NewEmail()
	//设置发送方的邮箱
	e.From = from
	// 设置接收方的邮箱
	e.To = []string{to}
	e.Subject = content
	//设置文件发送的内容
	e.Text = []byte(content)
	//设置服务器相关的配置
	err := e.Send(smtpAddress+":"+smtpPort, smtp.PlainAuth("", from, passcode, smtpAddress))
	if err != nil {
		log.Fatal(err)
	}
}

func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func main() {
	godotenv.Load()
	launcher.DefaultBrowserDir = "./chromium"
	page := rod.New().MustConnect().MustPage("")

	println(1)
	router := page.HijackRequests()
	// 阻止这个链接的加载，因为会判断权限，让页面跳转到广东统一身份认证平台
	router.MustAdd("*/sq-utils.js", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return
	})

	go router.Run()
	println(2)
	page.Timeout(60 * time.Second).MustNavigate("https://crj.gdga.gd.gov.cn/gdfwzww/views/jdcx/jdcxjg.html").MustWaitLoad()
	println(3)
	page.Timeout(60 * time.Second).MustElement("#ZJHM").MustInput(os.Getenv("ID"))
	println(4)
	page.Timeout(60 * time.Second).MustElement("body div.gd-form-item.table-wsyymlpt button").MustClick()
	println(5)
	statusDOM := page.Timeout(60 * time.Second).MustElement("#query_search_table div.col-sm-2.states")
	println(6)
	text := statusDOM.Timeout(60 * time.Second).MustText()
	println(7)
	println("状态：", text)

	status := os.Getenv("STATUS")
	if status == "" {
		status = "正在审批中"
	}

	if text != "" && text != status {
		sendMail(text)
	}
}
