package framework

import (
	"log"
	"math/rand"
	"time"

	bot "github.com/astralservices/bots/packages/commands/bot"
	fun "github.com/astralservices/bots/packages/commands/fun"
	moderation "github.com/astralservices/bots/packages/commands/moderation"
	"github.com/astralservices/bots/packages/middlewares"
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Bot struct {
	Bot     utils.IBot
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

	handler := shireikan.New(&shireikan.Config{
		GeneralPrefix:         i.Bot.Settings.Prefix,
		AllowBots:             false,
		AllowDM:               false,
		ExecuteOnEdit:         true,
		InvokeToLower:         true,
		UseDefaultHelpCommand: false,
		OnError: func(ctx shireikan.Context, typ shireikan.ErrorType, err error) {
			utils.ErrorHandler(err)
		},
	})

	// Register middlewares
	handler.Register(&middlewares.Bot{Settings: i.Bot})
	handler.Register(&middlewares.PermissionsMiddleware{})

	// Register commands
	handler.Register(&bot.Ping{})
	handler.Register(&bot.Help{})
	handler.Register(&bot.Region{})
	handler.Register(&bot.Info{})

	handler.Register(&fun.Eightball{})
	// any reddit-related commands need to be removed for now
	// handler.Register(&fun.Cat{})

	handler.Register(&moderation.Ban{})

	// Setup command handler
	handler.Setup(s)
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
