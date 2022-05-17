package main

import (
	"flag"
	"log"
	"net/smtp"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jordan-wright/email"
)

func sendMail(content string) {
	var _, from, to, passcode, smtpAddress, smtpPort, _ = loadConfig()
	println("准备发送邮件...")

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

func loadConfig() (string, string, string, string, string, string, string) {
	var id, from, to, passcode, smtpAddress, smtpPort, status string
	flag.StringVar(&id, "id", "", "身份证")
	flag.StringVar(&from, "from", "", "发送方邮箱")
	flag.StringVar(&to, "to", "", "接收方邮箱")
	flag.StringVar(&passcode, "passcode", "", "邮箱登录授权码/密码")
	flag.StringVar(&smtpAddress, "smtp", "", "smtp服务，不带端口")
	flag.StringVar(&smtpPort, "smtp-port", "", "smtp端口")
	flag.StringVar(&status, "status", "", "用来做对比的状态字符")

	flag.Parse()

	if status == "" {
		status = "正在审批中"
	}
	if smtpAddress == "" {
		smtpAddress = "smtp.qq.com"
	}
	if smtpPort == "" {
		smtpPort = "25"
	}

	println(id, from, to, passcode, smtpAddress, smtpPort, status)

	return id, from, to, passcode, smtpAddress, smtpPort, status
}

func main() {
	var id, _, _, _, _, _, status = loadConfig()
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
	page.Timeout(60 * time.Second).MustElement("#ZJHM").MustInput(id)
	println(4)
	page.Timeout(60 * time.Second).MustElement("body div.gd-form-item.table-wsyymlpt button").MustClick()
	println(5)
	statusDOM := page.Timeout(60 * time.Second).MustElement("#query_search_table div.col-sm-2.states")
	println(6)
	text := statusDOM.Timeout(60 * time.Second).MustText()
	println(7)
	println("状态：", text)

	if text != "" && text != status {
		sendMail(text)
	}
}
