package main

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
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
	email.SetBody(mail.TextHTML, message.Content)

	err = email.Send(client)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	app.InfoLog.Println("Email sent!")
}
