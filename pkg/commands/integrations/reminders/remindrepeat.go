package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"github.com/carlescere/scheduler"
)

var RemindRepeatCommand = &dgc.Command{
	Name:        "remindrepeat",
	Description: "Remind yourself or someone else of something at a set interval",
	Domain:      "astral.integrations.reminders",
	Usage:       "remindrepeat [@user or me] [time] [message]",
	Example:     "remindrepeat me 1h to do something",
	Category:    "Reminders",
	Aliases:     []string{"remindrepeat", "reminderrepeat", "rremind", "remindr"},
	Slash:       true,
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

		wi, err := integrations.GetWorkspaceIntegrationForCommand(ctx, ReminderIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		iD, err := database.GetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wi.ID)

		if err != nil {
			// do nothing because the user may not have any data
		}

		var reminders []types.Reminder

		re := types.Reminder{
			Time:           t,
			Msg:            msg,
			UserID:         user,
			Repeating:      true,
			RepeatInterval: duration.String(),
			MessageID:      ctx.Event.ID,
			CreatedAt:      time.Now(),
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

		err = database.SetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wi.ID, types.ReminderIntegrationData{
			Reminders: reminders,
		})

		// Create the reminder

		notifier := &integrations.DiscordNotifier{
			UserID:   user,
			Session:  ctx.Session,
			Reminder: re,
			WiID:     wi.ID,
		}

		job, err := scheduler.Every(int(duration.Seconds())).Seconds().NotImmediately().Run(func() {
			notifier.Notify(user, msg)
		})

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		ctx.CustomObjects.Set(fmt.Sprintf("job-%s", ctx.Event.ID), job)

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
			Description: fmt.Sprintf("Okay! I'll remind you every `%s` to %s starting <t:%d>", duration.String(), msg, t.Unix()),
			Color:       0x00ff00,
			Fields:      fields,
		}))

	},
}
