package main

import (
	"codeberg.org/Yonle/go-wrsbmkg"
	"codeberg.org/Yonle/go-wrsbmkg/helper"
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"net/http"
	"strconv"
	"time"
)

var WIB = time.FixedZone("WIB", +7*60*60)

func sendWarning(ctx context.Context, b *bot.Bot, shakemapURL string, msg string) {
	for {
		_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:    config.ChatID,
			Caption:   msg,
			ParseMode: models.ParseModeMarkdownV1,
			Photo:     &models.InputFileString{Data: shakemapURL},
		})

		if err != nil {
			time.Sleep(time.Second*15)
			continue
		}

		break
	}
}

func startBMKG(ctx context.Context, b *bot.Bot) {
	p := wrsbmkg.Penerima{
		Gempa:    make(chan wrsbmkg.DataJSON),
		Realtime: make(chan wrsbmkg.DataJSON),
		Narasi:   make(chan string),

		Interval: time.Second * 15,
		API_URL:  wrsbmkg.DEFAULT_API_URL,

		HTTP_Client: http.Client{
			Timeout: time.Second * 30,
		},
	}

	p.MulaiPolling(ctx)

listener:
	for {
		select {
		case g := <-p.Gempa:
			gempa := helper.ParseGempa(g)
			msg := fmt.Sprintf(
				"*%s*\n\n%s\n\n%s\n\n%s\n\n%s\n",
				gempa.Subject,
				gempa.Description,
				gempa.Area,
				gempa.Potential,
				gempa.Instruction,
			)

			shakemapURL := "https://bmkg-content-inatews.storage.googleapis.com/" + gempa.Shakemap

			go sendWarning(ctx, b, shakemapURL, msg)
		case r := <-p.Realtime:
			realtime := helper.ParseRealtime(r)
			t, _ := time.Parse(time.DateTime, realtime.Time)
			tl := t.In(WIB)
			date := tl.Format(time.DateOnly)
			ft := tl.Format(time.Kitchen)
			msg := fmt.Sprintf(
				"*%s*\n"+
					"`"+
					"Tanggal   : %s\n"+
					"Waktu     : %s\n"+
					"Magnitudo : M%.1f\n"+
					"Kedalaman : %.1f KM\n"+
					"Fase      : %v\n"+
					"Status    : %s"+
					"`",
				realtime.Place,
				date,
				ft,
				realtime.Magnitude,
				realtime.Depth,
				realtime.Phase,
				realtime.Status,
			)

			lat, _ := strconv.ParseFloat(realtime.Coordinates[1].(string), 64)
			long, _ := strconv.ParseFloat(realtime.Coordinates[0].(string), 64)

			venueTitle := fmt.Sprintf("M%.1f, %s %s", realtime.Magnitude, date, ft)

			m, err := b.SendVenue(ctx, &bot.SendVenueParams{
				ChatID:              config.ChatID,
				DisableNotification: true,
				Latitude:            lat,
				Longitude:           long,
				Title:               realtime.Place,
				Address:             venueTitle,
			})

			if err != nil {
				fmt.Println(err)
				continue listener
			}

			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:              config.ChatID,
				Text:                msg,
				ParseMode:           models.ParseModeMarkdownV1,
				DisableNotification: true,
				ReplyParameters:     &models.ReplyParameters{MessageID: m.ID},
			}); err != nil {
				fmt.Println(err)
				continue listener
			}
		case n := <-p.Narasi:
			narasi := helper.CleanNarasi(n)

			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    config.ChatID,
				Text:      narasi,
			})

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
