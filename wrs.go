package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"codeberg.org/Yonle/go-wrsbmkg"
	"codeberg.org/Yonle/go-wrsbmkg/helper"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var WIB = time.FixedZone("WIB", +7*60*60)
var currentEventID string
var narasi = make(chan string)

func startBMKG(ctx context.Context, b *bot.Bot) {
	p := wrsbmkg.BuatPenerima()

	p.MulaiPolling(ctx)

listener:
	for {
		select {
		case g := <-p.Gempa:
			gempa := helper.ParseGempa(g)
			currentEventID = gempa.EventID

			if config.MinMag >= gempa.Magnitude {
				continue listener
			}

			go func() {
				teksNarasi, err := p.FetchNarasi(ctx, gempa.EventID, time.Now().Add(48*time.Hour))
				if err != nil {
					return
				}

				if IsNewMessage(gempa.EventID + "-Narasi") {
					narasi <- teksNarasi
				}
			}()

			if !IsNewMessage(gempa.Identifier) {
				continue listener
			}

			msg := fmt.Sprintf(
				"*%s*\n\n%s\n\n%s\n\n%s\n\n%s\n",
				gempa.Subject,
				gempa.Description,
				gempa.Area,
				gempa.Potential,
				gempa.Instruction,
			)

			log.Printf("wrs: Got event ID: %s", gempa.EventID)
			log.Printf(gempa.Headline)

			// send headline first. As the shakemap isn't really ready at the time of the incident.
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: config.ChatID,
				Text:   gempa.Headline,
			}); err != nil {
				log.Printf("bot: Failed to send headline: %s", err)
			}

			go sendPhoto(ctx, b, gempa.Shakemap, msg)

			// Below code is for handling Tsunami warning.
			go sendPhoto(ctx, b, gempa.WZMap, "")
			go sendPhoto(ctx, b, gempa.TTMap, "")
			go sendPhoto(ctx, b, gempa.SSHMap, "")

			var zonaObservasiText string

			for _, area := range gempa.ObsAreas {
				zonaObservasiText += fmt.Sprintf(
					"- *%s* (%s %s) dengan ketinggian *%s Meter* pada tanggal *%s* pukul *%s*\n",
					area.Location, area.Latitude, area.Longitude, area.Height, area.Date, area.Time,
				)
			}

			if len(zonaObservasiText) > 0 {
				headerText := "*TELAH TERJADI GEMPABUMI BERPOTENSI TSUNAMI*\n\n"
				headerText += fmt.Sprintf(
					"Gempa terjadi pada *%s*, Pukul *%s*, berkekuatan *M%.1f*, dengan kedalaman *%s* pada jarak *%s*\n",
					gempa.Date, gempa.Time, gempa.Magnitude, gempa.Depth, gempa.Area,
				)

				headerText += "\nBerdasarkan pengamatan muka air laut, tsunami telah terdeteksi di wilayah berikut:"

				zonaObservasiText = headerText + "\n" + zonaObservasiText

				fmt.Println("\n---\n" + zonaObservasiText)

				if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    config.ChatID,
					Text:      zonaObservasiText,
					ParseMode: models.ParseModeMarkdownV1,
				}); err != nil {
					log.Printf("bot: Failed to send zonaObservasiText: %s", err)
				}

			}

			var zonaPeringatanText string

			for _, area := range gempa.WZAreas {
				zonaPeringatanText += fmt.Sprintf(
					"- *%s*: *%s, %s* (estimasi waktu tiba: *%s %s*)\n",
					area.Level,
					area.Province,
					area.District,
					area.Date,
					area.Time,
				)
			}

			if len(zonaPeringatanText) > 0 {
				zonaPeringatanText += fmt.Sprintf(
					"\n*Saran dan Arahan Status Peringatan*\n1. %s\n2. %s\n3. %s",
					gempa.Instruction1,
					gempa.Instruction2,
					gempa.Instruction3,
				)

				zonaPeringatanText = "Daerah yang berpotensi tsunami berdasarkan pemodelan:\n" + zonaPeringatanText

				fmt.Println(zonaPeringatanText)

				if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    config.ChatID,
					Text:      zonaPeringatanText,
					ParseMode: models.ParseModeMarkdownV1,
				}); err != nil {
					log.Printf("bot: Failed to send zonaPeringatanText: %s", err)
				}
			}
		case r := <-p.Realtime:
			realtime := helper.ParseRealtime(r)

			if config.MinMag >= realtime.Magnitude {
				continue listener
			}

			if !IsNewMessage(realtime.Time) {
				continue listener
			}

			if !checkFilter(realtime.Place) {
				continue listener
			}

			t, _ := time.Parse(time.DateTime, realtime.Time)
			tl := t.In(WIB)
			date := tl.Format(time.DateOnly)
			ft := tl.Format(time.Kitchen)
			msg := fmt.Sprintf(
				"*M%.1f* %s\n"+
					"`"+
					"Tanggal   : %s\n"+
					"Waktu     : %s\n"+
					"Kedalaman : %.1f KM\n"+
					"Fase      : %v\n"+
					"Status    : %s"+
					"`",
				realtime.Magnitude,
				realtime.Place,
				date,
				ft,
				realtime.Depth,
				realtime.Phase,
				realtime.Status,
			)

			log.Printf("wrs: Got realtime info: M%.1f %s", realtime.Magnitude, realtime.Place)

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
				log.Printf("bot: Failed to send venue: %s", err)
				continue listener
			}

			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:              config.ChatID,
				Text:                msg,
				ParseMode:           models.ParseModeMarkdownV1,
				DisableNotification: true,
				ReplyParameters:     &models.ReplyParameters{MessageID: m.ID},
			}); err != nil {
				log.Printf("bot: Failed to send realtime info message: %s", err)
				continue listener
			}
		case n := <-narasi:
			narasi := helper.CleanNarasi(n)
			log.Println("wrs: Got narasi")

			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: config.ChatID,
				Text:   narasi,
			})

			if err != nil {
				log.Printf("bot: Failed to send narasi: %s", err)
			}
		}
	}
}
