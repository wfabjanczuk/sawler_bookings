package main

import (
	"fmt"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"io/ioutil"
	"strings"
	"time"
)

func listenForMail() {
	go func() {
		for {
			message := <-app.MailChannel
			sendMessage(message)
		}
	}()
}

func sendMessage(message models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	email := mail.NewMSG()
	email.SetFrom(message.From).AddTo(message.To).SetSubject(message.Subject)

	if message.Template == "" {
		email.SetBody(mail.TextHTML, message.Content)
	} else {
		data, err := ioutil.ReadFile(fmt.Sprintf("./templates/emails/%s", message.Template))
		if err != nil {
			app.ErrorLog.Println(err)
			return
		}

		message.Content = string(data)
		message.Content = strings.Replace(message.Content, "[%title%]", message.TemplateTitle, 1)
		message.Content = strings.Replace(message.Content, "[%body%]", message.TemplateBody, 1)
		email.SetBody(mail.TextHTML, message.Content)
	}

	err = email.Send(client)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	app.InfoLog.Println("Email sent!")
}
