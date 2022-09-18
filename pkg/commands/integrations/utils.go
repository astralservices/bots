package integrations

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KrishanBhalla/reminder"
	"github.com/KrishanBhalla/reminder/schedule"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"github.com/carlescere/scheduler"
)

type DiscordNotifier struct {
	Session  *discordgo.Session
	Reminder types.Reminder
	WiID     int
	UserID   string
}

func (d *DiscordNotifier) Notify(title, message string) error {
	c, err := d.Session.UserChannelCreate(d.UserID)

	if err != nil {
		return err
	}

	valueStr := fmt.Sprintf("You asked me to remind you about this at <t:%d>.", d.Reminder.Time.Unix())

	if d.Reminder.Repeating {
		valueStr = fmt.Sprintf("You asked me to remind you about this every `%s` starting <t:%d>.", d.Reminder.RepeatInterval, d.Reminder.CreatedAt.Unix())
	}

	_, err = d.Session.ChannelMessageSendComplex(c.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Reminder 👆",
				Description: message,
				Color:       0x00ff00,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Why am I getting this?",
						Value: valueStr,
					},
				},
			},
		},
	})

	if !d.Reminder.Repeating {
		database := db.New()

		iD, err := database.GetIntegrationDataForUser(d.UserID, "3dc87d39-a037-48fa-85b0-0243e6593883", d.WiID)

		if err != nil {
			return err
		}

		var reminders types.ReminderIntegrationData

		jsonStr, err := json.Marshal(iD.Data)

		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonStr, &reminders)

		if err != nil {
			return err
		}

		for i, r := range reminders.Reminders {
			if r.MessageID == d.Reminder.MessageID {
				reminders.Reminders = append(reminders.Reminders[:i], reminders.Reminders[i+1:]...)
			}
		}

		err = database.SetIntegrationDataForUser(d.UserID, "3dc87d39-a037-48fa-85b0-0243e6593883", d.WiID, types.ReminderIntegrationData{
			Reminders: reminders.Reminders,
		})

		if err != nil {
			return err
		}

		return nil
	}

	return err
}

func GetWorkspaceIntegrationForCommand(ctx *dgc.Ctx, integrationID string) (workspaceIntegration types.WorkspaceIntegration, err error) {
	data := ctx.CustomObjects.MustGet("workspaceIntegrations")

	workspaceIntegrations := data.([]types.WorkspaceIntegration)

	for _, workspaceIntegration := range workspaceIntegrations {
		if workspaceIntegration.Integration == integrationID {
			return workspaceIntegration, nil
		}
	}

	return types.WorkspaceIntegration{}, fmt.Errorf("no workspace integration found for integration %s", integrationID)
}

func SetupReminders(session *discordgo.Session, self types.Bot) (err error) {
	database := db.New()

	d, err := database.GetIntegrationDataForWorkspace(*self.Workspace, "3dc87d39-a037-48fa-85b0-0243e6593883")

	if err != nil {
		return err
	}

	for _, data := range d {
		jsonStr, err := json.Marshal(data.Data)

		if err != nil {
			return err
		}

		var reminders types.ReminderIntegrationData

		err = json.Unmarshal(jsonStr, &reminders)

		if err != nil {
			return err
		}

		for _, r := range reminders.Reminders {
			s, _ := schedule.NewSchedule(time.RFC1123Z, r.Time.Format(time.RFC1123Z), "UTC")

			notifier := &DiscordNotifier{
				Session:  session,
				UserID:   r.UserID,
				Reminder: r,
				WiID:     d[0].WorkspaceIntegration,
			}

			if r.Repeating {
				// convert to duration
				d, err := time.ParseDuration(r.RepeatInterval)

				if err != nil {
					return err
				}

				scheduler.Every(int(d.Seconds())).Seconds().NotImmediately().Run(func() {
					notifier.Notify(r.UserID, r.Msg)
				})
			} else {
				rem := reminder.Reminder{
					Schedule: s,
					Notifier: notifier,
				}

				go rem.Remind(r.UserID, r.Msg)
			}
		}
	}

	return nil
}
