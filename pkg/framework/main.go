package framework

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	bot "github.com/astralservices/bots/pkg/commands/bot"
	fun "github.com/astralservices/bots/pkg/commands/fun"
	"github.com/astralservices/bots/pkg/commands/integrations"
	college "github.com/astralservices/bots/pkg/commands/integrations/college"
	lastfm "github.com/astralservices/bots/pkg/commands/integrations/lastfm"
	mcbroken "github.com/astralservices/bots/pkg/commands/integrations/mcbroken"
	reactionroles "github.com/astralservices/bots/pkg/commands/integrations/reaction_roles"
	reminders "github.com/astralservices/bots/pkg/commands/integrations/reminders"
	moderation "github.com/astralservices/bots/pkg/commands/moderation"
	"github.com/astralservices/bots/pkg/commands/utility"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/middlewares"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Bot     types.Bot
	Session *discordgo.Session

	statusInterval chan int

	analyticsCache  types.BotAnalytics
	analyticsSync   chan bool
	analyticsTicker *time.Ticker

	permissionsMiddleware *middlewares.PermissionsMiddleware
}

func (i *Bot) Initialize() {
	log.Println("Initialize", *i.Bot.ID)
	rand.Seed(time.Now().Unix() + i.Bot.CreatedAt.Unix())

	s, err := discordgo.New("Bot " + i.Bot.Token)

	// Intents

	s.Identify.Intents |= discordgo.IntentsGuildMessages
	s.Identify.Intents |= discordgo.IntentsGuildMessageReactions
	s.Identify.Intents |= discordgo.IntentsGuildMembers

	if err != nil {
		utils.ErrorHandler(err)
		log.Println("Error creating DiscordGo session, destroying bot")
		i.Destroy()
		return
	}

	err = s.Open()
	if err != nil {
		utils.ErrorHandler(err)
		log.Println("Could not open connection, destroying bot")
		i.Destroy()
		return
	}

	i.Session = s

	i.setStatus()
	go i.updateStatusLoop()

	router := dgc.Create(&dgc.Router{
		Prefixes: []string{i.Bot.Settings.Prefix},
	})

	database := db.New()

	botMiddleware := middlewares.Bot{Bot: i.Bot}
	permissionsMiddleware := middlewares.PermissionsMiddleware{Bot: i.Bot}
	i.permissionsMiddleware = &permissionsMiddleware

	router.RegisterMiddleware(botMiddleware.BotMiddleware)
	router.RegisterMiddleware(i.analyticsMiddleware)
	router.RegisterMiddleware(permissionsMiddleware.Handle)

	///// BOT COMMANDS /////
	router.RegisterCmd(bot.Ping)
	router.RegisterCmd(bot.Help)
	router.RegisterCmd(bot.Info)
	router.RegisterCmd(bot.Region)

	///// FUN COMMANDS /////
	router.RegisterCmd(fun.Eightball)
	router.RegisterCmd(fun.Cat)
	router.RegisterCmd(fun.Dog)
	router.RegisterCmd(fun.Meme)
	router.RegisterCmd(fun.Rat)

	///// MODERATION /////
	router.RegisterCmd(moderation.BanCommand)
	router.RegisterCmd(moderation.HistoryCommand)
	router.RegisterCmd(moderation.CaseCommand)
	router.RegisterCmd(moderation.KickCommand)
	router.RegisterCmd(moderation.MuteCommand)
	router.RegisterCmd(moderation.UnmuteCommand)

	///// UTILITY /////
	router.RegisterCmd(utility.ServerInfoCommand)
	router.RegisterCmd(utility.StatsCommand)
	router.RegisterCmd(utility.WhoisCommand)

	///// INTEGRATIONS /////
	/// Register commands ///
	router.RegisterCmd(college.DormCommand)
	router.RegisterCmd(college.DormlistCommand)
	router.RegisterCmd(college.SetDormCommand)
	router.RegisterCmd(college.VerifyStudentCommand)

	router.RegisterCmd(lastfm.ScrobblesCommand)

	router.RegisterCmd(mcbroken.McBrokenCommand)

	router.RegisterCmd(reminders.RemindCommand)
	router.RegisterCmd(reminders.RemindRepeatCommand)
	router.RegisterCmd(reminders.ListRemindersCommand)
	router.RegisterCmd(reminders.DeleteReminderCommand)
	err = integrations.SetupReminders(i.Session, i.Bot)

	if err != nil {
		utils.ErrorHandler(fmt.Errorf("Error setting up reminders: %w", err))
	}

	router.RegisterCmd(reactionroles.ReactionRoleCommand)

	/// Register middleware ///
	workspaceIntegrations, err := database.GetIntegrationsForWorkspace(*i.Bot.Workspace)

	if err != nil {
		utils.ErrorHandler(err)
		panic(err)
	}

	for _, integration := range workspaceIntegrations {
		if integration.Integration == reactionroles.ReactionRolesIntegrationID {
			if integration.Enabled {
				data, err := database.GetIntegrationDataForWorkspace(*i.Bot.Workspace, integration.Integration)

				if err != nil {
					utils.ErrorHandler(err)
					panic(err)
				}

				jsonStr, err := json.Marshal(data[0].Data)

				if err != nil {
					utils.ErrorHandler(err)
					panic(err)
				}

				var reactionRolesData reactionroles.ReactionRolesData

				err = json.Unmarshal(jsonStr, &reactionRolesData)

				if err != nil {
					utils.ErrorHandler(err)
					panic(err)
				}

				s.AddHandler(reactionroles.HandleReactionRolesAdd(*i.Bot.Workspace))
				s.AddHandler(reactionroles.HandleReactionRolesRemove(*i.Bot.Workspace))
			}
		}
	}

	router.RegisterMiddleware(func(next dgc.ExecutionHandler) dgc.ExecutionHandler {
		return func(ctx *dgc.Ctx) {
			ctx.CustomObjects.Set("workspaceIntegrations", workspaceIntegrations)

			if ctx.Command.IntegrationID != "" {
				for _, integration := range workspaceIntegrations {
					if integration.Integration == ctx.Command.IntegrationID {
						if integration.Enabled {
							next(ctx)
						} else {
							ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
								Title:       "Integration Disabled",
								Description: "This integration is disabled for this workspace.",
								Color:       0xFF0000,
							}))
							return
						}
					}
				}
			} else {
				next(ctx)
			}
		}
	})

	router.Initialize(s)

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}
		i.analyticsCache.Messages++

		guild, err := s.State.Guild(m.GuildID)

		if err != nil {
			utils.ErrorHandler(err)
		}

		i.analyticsCache.Members = guild.ApproximateMemberCount
	})

	i.analyticsTicker = time.NewTicker(time.Minute * 10)
	i.analyticsCache = types.BotAnalytics{
		Commands: make(map[string]int),
	}

	go i.updateAnalyticsLoop()

	i.checkExpiredReports()
	go i.checkExpiredReportsLoop()

	guildStateTicker := time.NewTicker(time.Minute * 1)

	go func() {
		for {
			select {
			case <-guildStateTicker.C:
				guild, err := s.GuildWithCounts(i.Bot.Settings.Guild)

				if err != nil {
					utils.ErrorHandler(err)
					return
				}
				
				if guild == nil {
					return	
				}

				s.State.GuildAdd(guild)
			}
		}
	}()
}

func (i *Bot) Destroy() error {
	if i.Session == nil {
		return nil
	}
	if err := i.Session.Close(); err != nil {
		utils.ErrorHandler(err)
		return err
	} else {
		return nil
	}
}

func (i *Bot) Update() {
	i.permissionsMiddleware.UpdateConfig(&i.Bot)

	select {
	case i.statusInterval <- i.Bot.Settings.ActivityInterval:
	default:
	}
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
			currentActivityIndex := i.Bot.Settings.CurrentActivity

			if currentActivityIndex == -1 {
				currentActivityIndex = 0
			} else {
				currentActivityIndex = currentActivityIndex + 1
				if currentActivityIndex >= len(activities) {
					currentActivityIndex = 0
				}
			}

			selectedActivity = activities[currentActivityIndex]

			i.Bot.Settings.CurrentActivity = currentActivityIndex

			database := db.New()

			err := database.SetBot(*i.Bot.ID, i.Bot)

			if err != nil {
				utils.ErrorHandler(err)
			}
		}
	}

	i.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			selectedActivity,
		},
		Status: i.Bot.Settings.Status,
	})
}

func (i *Bot) analyticsMiddleware(next dgc.ExecutionHandler) dgc.ExecutionHandler {
	return func(ctx *dgc.Ctx) {
		command := ctx.Command.Domain

		if _, ok := i.analyticsCache.Commands[command]; ok {
			i.analyticsCache.Commands[command]++
		} else {
			if i.analyticsCache.Commands == nil {
				i.analyticsCache.Commands = make(map[string]int)
			}
			i.analyticsCache.Commands[command] = 1
		}

		next(ctx)
	}
}

func (i *Bot) updateAnalyticsLoop() {
	for {
		select {
		case <-i.analyticsSync:
			i.analyticsTicker.Stop()
			return

		case <-i.analyticsTicker.C:
			i.updateAnalytics()

		}
	}
}

func (i *Bot) updateAnalytics() {
	database := db.New()

	log.Printf("Updating analytics for %s", *i.Bot.ID)

	database.Supabase.DB.Rpc("commands_inc", "", map[string]interface{}{
		"command": i.analyticsCache.Commands,
		"row_id":  i.Bot.ID,
	})

	database.Supabase.DB.Rpc("messages_inc", "", map[string]interface{}{
		"x":      i.analyticsCache.Messages,
		"row_id": i.Bot.ID,
	})

	if i.analyticsCache.Members > 0 {
		database.Supabase.DB.Rpc("members_inc", "", map[string]interface{}{
			"x":      i.analyticsCache.Members,
			"row_id": i.Bot.ID,
		})
	}

	i.analyticsCache = types.BotAnalytics{
		Commands: make(map[string]int),
		Messages: 0,
		Members:  i.analyticsCache.Members,
	}
}

func (i *Bot) checkExpiredReports() {
	database := db.New()

	reports, err := database.GetReportsFiltered(types.ReportFilter{
		Expired: true,
		Bot:     *i.Bot.ID,
	})

	if err != nil {
		utils.ErrorHandler(err)
		return
	}

	for _, report := range reports {
		switch report.Action {
		case "ban":
			{
				err := i.Session.GuildBanDelete(report.Guild, report.User)

				if err != nil {
					utils.ErrorHandler(err)
				}
			}

		case "mute":
			{
				// find the muted role by name
				// remove the role from the user
				// if the user is muted through discord, remove the timeout

				guild, err := i.Session.Guild(report.Guild)

				if err != nil {
					utils.ErrorHandler(err)
					return
				}

				// find the muted role
				var mutedRole *discordgo.Role

				for _, role := range guild.Roles {
					if strings.Contains(strings.ToLower(role.Name), "muted") {
						mutedRole = role
						break
					}
				}

				// get the user
				victim, err := i.Session.GuildMember(report.Guild, report.User)

				if err != nil {
					utils.ErrorHandler(err)
					return
				}

				// remove the role, if the user has it
				for _, role := range victim.Roles {
					if role == mutedRole.ID {
						err := i.Session.GuildMemberRoleRemove(report.Guild, victim.User.ID, mutedRole.ID)

						if err != nil {
							utils.ErrorHandler(err)
							return
						}
					}
				}

				// remove the timeout, if the user has one
				if victim.CommunicationDisabledUntil != nil {
					err := i.Session.GuildMemberTimeout(report.Guild, victim.User.ID, nil)

					if err != nil {
						utils.ErrorHandler(err)
						return
					}
				}
			}
		}

		log.Println("Removing expired report", report.ID)
		database.ExpireReport(*report.ID)
	}
}

func (i *Bot) checkExpiredReportsLoop() {
	for range time.Tick(time.Minute) {
		i.checkExpiredReports()
	}
}
