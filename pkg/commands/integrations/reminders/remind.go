package integrations

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KrishanBhalla/reminder"
	"github.com/KrishanBhalla/reminder/schedule"
	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var ReminderIntegrationID = "3dc87d39-a037-48fa-85b0-0243e6593883"

var RemindCommand = &dgc.Command{
	Name:          "remind",
	Description:   "Remind yourself or someone else of something",
	Usage:         "remind [@user or me] [time] [message]",
	Example:       "remind me 1h to do something",
	Category:      "Reminders",
	IntegrationID: ReminderIntegrationID,
	Domain:        "astral.integrations.reminders",
	Aliases:       []string{"remind", "reminder"},
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		// Get the time
		duration, err := utils.ParseDuration(ctx.Arguments.Get(1).Raw())

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		t := time.Now().Add(duration)

		// Get the message
		a := ctx.Arguments.GetAll()[2:]

		msgArr := []string{}

		for _, v := range a {
			msgArr = append(msgArr, v.Raw())
		}

		msg := strings.Join(msgArr, " ")

		// Get the user
		user := ctx.Arguments.Get(0).Raw()

		if user == "me" {
			user = ctx.Event.Author.ID
		} else {
			user = ctx.Arguments.Get(0).AsUserMentionID()
		}

		// Add to the database
		database := db.New()

		wiID, err := integrations.GetWorkspaceIntegrationForCommand(ctx, ReminderIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		iD, err := database.GetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wiID)

		if err != nil {
			// do nothing because the user may not have any data
		}

		var reminders []types.Reminder

		re := types.Reminder{
			Time:      t,
			Msg:       msg,
			UserID:    user,
			Repeating: false,
			MessageID: ctx.Event.ID,
			CreatedAt: time.Now(),
		}

		if iD.Data != nil {
			jsonStr, err := json.Marshal(iD.Data)

			if err != nil {
				ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
				return
			}

			var d types.ReminderIntegrationData

			err = json.Unmarshal(jsonStr, &d)

			reminders = append(d.Reminders, re)
		} else {
			reminders = []types.Reminder{
				re,
			}
		}

		err = database.SetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wiID, types.ReminderIntegrationData{
			Reminders: reminders,
		})

		// Create the reminder

		s, _ := schedule.NewSchedule(time.RFC1123Z, t.Format(time.RFC1123Z), "Local")

		notifier := &integrations.DiscordNotifier{
			UserID:   user,
			Session:  ctx.Session,
			Reminder: re,
			WiID:     wiID,
		}

		rem := reminder.Reminder{
			Schedule: s,
			Notifier: notifier,
		}

		go rem.Remind(user, msg)

		var fields []*discordgo.MessageEmbedField

		if err != nil {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Error adding to database",
				Value:  fmt.Sprintf("%s\n\nYour reminder is still set, however if the bot restarts, it will be lost.", err.Error()),
				Inline: false,
			})
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Reminder set!",
			Description: fmt.Sprintf("I'll remind you <t:%d:R>", t.Unix()),
			Color:       0x00ff00,
			Fields:      fields,
		}))
	},
}
