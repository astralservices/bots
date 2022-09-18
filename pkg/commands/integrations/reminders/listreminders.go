package integrations

import (
	"encoding/json"
	"fmt"

	"github.com/astralservices/bots/pkg/commands/integrations"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var ListRemindersCommand = &dgc.Command{
	Name:        "listreminders",
	Description: "List all of your reminders",
	Domain:      "astral.integrations.reminders",
	Aliases:     []string{"listreminders", "listreminder", "listremind", "reminders", "reminderls", "remindls"},
	Category:    "Reminders",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		wi, err := integrations.GetWorkspaceIntegrationForCommand(ctx, ReminderIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		iD, err := database.GetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wi.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		var reminders types.ReminderIntegrationData

		jsonStr, err := json.Marshal(iD.Data)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		err = json.Unmarshal(jsonStr, &reminders)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Reminders",
			Description: "Here are all of your reminders",
			Color:       0x00ff00,
			Fields:      []*discordgo.MessageEmbedField{},
		}

		for _, r := range reminders.Reminders {
			valText := fmt.Sprintf("**Reminder:** %s\n**Time:** <t:%d>", r.Msg, r.Time.Unix())

			if r.Repeating {
				valText = fmt.Sprintf("**Reminder:** %s\n**Every:** `%s`", r.Msg, r.RepeatInterval)
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("ID: %s", r.MessageID),
				Value:  valText,
				Inline: false,
			})
		}

		ctx.ReplyEmbed(embed)
	},
}
