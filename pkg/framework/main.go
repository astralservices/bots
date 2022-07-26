package framework

import (
	"log"
	"math/rand"
	"time"

	bot "github.com/astralservices/bots/pkg/commands/bot"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Bot     types.Bot
	Session *discordgo.Session

	statusInterval chan int
}

func (i *Bot) Initialize() {
	log.Println("Initialize", *i.Bot.ID)
	rand.Seed(time.Now().Unix() + i.Bot.CreatedAt.Unix())

	s, err := discordgo.New("Bot " + i.Bot.Token)

	if err != nil {
		utils.ErrorHandler(err)
	}

	err = s.Open()
	if err != nil {
		utils.ErrorHandler(err)
	}

	i.Session = s

	i.setStatus()
	go i.updateStatusLoop()

	router := dgc.Create(&dgc.Router{
		Prefixes: []string{i.Bot.Settings.Prefix},
	})

	router.InitializeStorage(*i.Bot.ID)
	router.Storage[*i.Bot.ID].Set("self", i.Bot)

	router.RegisterCmd(&bot.Ping)

	router.Initialize(s)
}

func (i *Bot) Destroy() error {
	if err := i.Session.Close(); err != nil {
		utils.ErrorHandler(err)
		return err
	} else {
		return nil
	}
}

func (i *Bot) Update() error {
	i.statusInterval <- i.Bot.Settings.ActivityInterval
	return nil
}

func (i *Bot) updateStatusLoop() {
	for range time.Tick(time.Second * time.Duration(<-i.statusInterval)) {
		i.setStatus()
	}
}

func (i *Bot) setStatus() {
	var activities []*discordgo.Activity

	for _, a := range i.Bot.Settings.Activities {
		activities = append(activities, &discordgo.Activity{
			Name: a.Name,
			Type: utils.ConvertStringToActivityType(a.Type),
		})
	}

	var selectedActivity *discordgo.Activity

	if len(activities) > 0 {
		if i.Bot.Settings.RandomizeActivities {
			selectedActivity = activities[rand.Intn(len(activities))]
		} else {
			selectedActivity = activities[i.Bot.Settings.CurrentActivity]
		}
	}

	i.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			selectedActivity,
		},
		Status: i.Bot.Settings.Status,
	})
}
