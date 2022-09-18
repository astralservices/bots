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
	"github.com/carlescere/scheduler"
)

var DeleteReminderCommand = &dgc.Command{
	Name:        "deletereminder",
	Description: "Delete a reminder",
	Domain:      "astral.integrations.reminders",
	Aliases:     []string{"deletereminder", "delreminder", "delremind", "rmreminder", "rmremind"},
	Category:    "Reminders",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		// Get the reminder id
		id := ctx.Arguments.Get(0).Raw()

		wi, err := integrations.GetWorkspaceIntegrationForCommand(ctx, ReminderIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		iD, err := database.GetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wi.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
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

		for i, r := range reminders.Reminders {
			if r.MessageID == id {
				reminders.Reminders = append(reminders.Reminders[:i], reminders.Reminders[i+1:]...)
			}
		}

		err = database.SetIntegrationDataForUser(ctx.Event.Author.ID, ReminderIntegrationID, wi.ID, types.ReminderIntegrationData{
			Reminders: reminders.Reminders,
		})

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		j, found := ctx.CustomObjects.Get(fmt.Sprintf("job-%s", id))

		if found {
			job := j.(*scheduler.Job)

			job.Quit <- true
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Reminder Deleted",
			Description: "Your reminder has been deleted. You may receive a notification if it's in the queue.",
			Color:       0x00ff00,
		}))
	},
}
