package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/astralservices/bots/packages/framework"
	"github.com/astralservices/bots/utils"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/nedpals/supabase-go"
	realtimego "github.com/overseedio/realtime-go"
)

type Cache struct {
	Bots map[string]*framework.Bot
}

func main() {
	err := godotenv.Load("./.env")

	if err != nil {
		utils.ErrorHandler(err)
	}

	ENDPOINT := os.Getenv("SUPABASE_URL")
	API_KEY := os.Getenv("SUPABASE_KEY")
	REGION := os.Getenv("REGION")

	if REGION == "" {
		REGION, err = os.Hostname()
		if err != nil {
			utils.ErrorHandler(err)
		}
	}

	cache := Cache{
		Bots: make(map[string]*framework.Bot),
	}

	go func() {
		botsRealtimeOpts := utils.RealtimeOptions{
			Endpoint: ENDPOINT,
			Key:      API_KEY,
			Table:    "bots",
			OnInsert: func(m realtimego.Message) {
				var payload utils.RealtimePayload[utils.IBot]

				str, err := json.Marshal(m.Payload)

				if err != nil {
					utils.ErrorHandler(err)
				}

				err = json.Unmarshal(str, &payload)
				if err != nil {
					utils.ErrorHandler(err)
				}

				if payload.Record.Region == REGION {
					cache.AddBot(payload.Record)
				}
			},
			OnDelete: func(m realtimego.Message) {
				var payload utils.RealtimePayload[utils.IBot]

				str, err := json.Marshal(m.Payload)

				if err != nil {
					utils.ErrorHandler(err)
				}

				err = json.Unmarshal(str, &payload)
				if err != nil {
					utils.ErrorHandler(err)
				}

				if payload.Record.Region == REGION {
					cache.DeleteBot(payload.Record)
				}
			},
			OnUpdate: func(m realtimego.Message) {
				var payload utils.RealtimePayload[utils.IBot]

				str, err := json.Marshal(m.Payload)

				if err != nil {
					utils.ErrorHandler(err)
				}

				err = json.Unmarshal(str, &payload)
				if err != nil {
					utils.ErrorHandler(err)
				}

				if payload.OldRecord.Region != payload.Record.Region {
					if payload.OldRecord.Region == REGION {
						cache.DeleteBot(payload.OldRecord)
					} else if payload.Record.Region == REGION {
						cache.AddBot(payload.Record)
					}
				}

				if payload.OldRecord.Settings.CurrentActivity != payload.Record.Settings.CurrentActivity {
					return // ignore activity changes
				}

				if payload.Record.Region == REGION {
					cache.UpdateBot(payload.Record)
				}
			},
		}

		botsRealtimeOpts.SetupRealtime()
	}()

	supabaseClient := supabase.CreateClient(ENDPOINT, API_KEY)

	sentry.Init(sentry.ClientOptions{
		Dsn: "https://681b6c2ca26a4c258e77a0068c84404f@gt.astralapp.io/4",
	})

	var bots []utils.IBot

	supabaseClient.DB.From("bots").Select("*").Eq("region", REGION).Execute(&bots)

	for _, bot := range bots {
		cache.AddBot(bot)
	}

	// prevent program from exiting
	defer func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc
	}()
}

func (c *Cache) AddBot(bot utils.IBot) {
	log.Println("Add Bot", *bot.ID)

	botClient := framework.Bot{
		Bot: bot,
	}

	botClient.Initialize()

	c.Bots[*bot.ID] = &botClient
}

func (c *Cache) DeleteBot(bot utils.IBot) {
	log.Println("Delete Bot", *bot.ID)

	err := c.Bots[*bot.ID].Destroy()

	if err != nil {
		utils.ErrorHandler(err)
	}

	delete(c.Bots, *bot.ID)
}

func (c *Cache) UpdateBot(bot utils.IBot) {
	log.Println("Update Bot", *bot.ID)

	c.Bots[*bot.ID].Bot = bot

	c.Bots[*bot.ID].Update()
}
