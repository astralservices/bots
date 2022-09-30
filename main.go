package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/framework"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	realtimego "github.com/overseedio/realtime-go"
)

type Cache struct {
	Bots map[string]*framework.Bot
}

func main() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load("./.env")

		if err != nil {
			utils.ErrorHandler(err)
		}
	}

	ENDPOINT := os.Getenv("SUPABASE_URL")
	API_KEY := os.Getenv("SUPABASE_KEY")
	REGION := os.Getenv("REGION")

	if REGION == "" {
		r, err := os.Hostname()
		REGION = r
		if err != nil {
			utils.ErrorHandler(err)
		}
	}

	database := db.New()

	cache := Cache{
		Bots: make(map[string]*framework.Bot),
	}

	c, err := realtimego.NewClient(ENDPOINT, API_KEY, realtimego.WithUserToken(API_KEY))

	if err != nil {
		panic(err)
	}

	// connect to server
	err = c.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// create and subscribe to channel
	db := "realtime"
	schema := "public"
	table := "bots"
	botsRt, err := c.Channel(realtimego.WithTable(&db, &schema, &table))
	if err != nil {
		log.Fatal(err)
	}

	otherTable := "workspace_integrations"
	integrationsRt, err := c.Channel(realtimego.WithTable(&db, &schema, &otherTable))

	// setup hooks
	botsRt.OnInsert = func(m realtimego.Message) {
		var payload utils.RealtimePayload[types.Bot]

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
	}
	botsRt.OnDelete = func(m realtimego.Message) {
		var payload utils.RealtimePayload[types.Bot]

		str, err := json.Marshal(m.Payload)

		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		err = json.Unmarshal(str, &payload)
		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		if payload.Record.Region == REGION {
			cache.DeleteBot(payload.Record)
		}
	}

	botsRt.OnUpdate = func(m realtimego.Message) {
		var payload utils.RealtimePayload[types.Bot]

		str, err := json.Marshal(m.Payload)

		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		err = json.Unmarshal(str, &payload)
		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		if payload.OldRecord.Settings.CurrentActivity != payload.Record.Settings.CurrentActivity {
			return // ignore activity changes
		}

		if payload.OldRecord.Region != payload.Record.Region {
			if payload.OldRecord.Region == REGION {
				cache.DeleteBot(payload.OldRecord)
				return
			} else if payload.Record.Region == REGION {
				cache.AddBot(payload.Record)
				return
			}
		}

		if payload.Record.Region == REGION {
			cache.UpdateBot(payload.Record)
		}
	}

	integrationsFunc := func(m realtimego.Message) {
		var payload utils.RealtimePayload[types.WorkspaceIntegration]

		str, err := json.Marshal(m.Payload)

		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		err = json.Unmarshal(str, &payload)
		if err != nil {
			utils.ErrorHandler(err)
			return
		}

		// find bot then restart
		bot, err := database.GetBotForWorkspace(payload.Record.Workspace)

		b := cache.Bots[*bot.ID]

		if b != nil {
			log.Println("Restart Bot", *bot.ID)

			cache.DeleteBot(b.Bot)
			cache.AddBot(b.Bot)
		}
	}

	integrationsRt.OnInsert = integrationsFunc
	integrationsRt.OnUpdate = integrationsFunc
	integrationsRt.OnDelete = integrationsFunc

	// subscribe to channel
	err = botsRt.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	err = integrationsRt.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	sentry.Init(sentry.ClientOptions{
		Dsn: "https://681b6c2ca26a4c258e77a0068c84404f@gt.astralapp.io/4",
	})

	bots, err := database.GetAllBotsForRegion(REGION)

	if err != nil {
		utils.ErrorHandler(err)
	}

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

func (c *Cache) AddBot(bot types.Bot) {
	log.Println("Add Bot", *bot.ID)

	botClient := framework.Bot{
		Bot: bot,
	}

	botClient.Initialize()

	c.Bots[*bot.ID] = &botClient
}

func (c *Cache) DeleteBot(bot types.Bot) {
	log.Println("Delete Bot", *bot.ID)

	err := c.Bots[*bot.ID].Destroy()

	if err != nil {
		utils.ErrorHandler(err)
	}

	delete(c.Bots, *bot.ID)
}

func (c *Cache) UpdateBot(bot types.Bot) {
	log.Println("Update Bot", *bot.ID)

	c.Bots[*bot.ID].Bot = bot

	c.Bots[*bot.ID].Update()
}
