package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/zekroTJA/shireikan"
)

type Help struct {
}

func (c *Help) GetInvokes() []string {
	return []string{"help", "h", "?", "man"}
}

func (c *Help) GetDescription() string {
	return "Displays command list or help of specific command"
}

func (c *Help) GetHelp() string {
	return "`help` - display command list\n" +
		"`help <command>` - display help of specific command"
}

func (c *Help) GetGroup() string {
	return utils.CategoryBot
}

func (c *Help) GetDomainName() string {
	return "astral.bot.help"
}

func (c *Help) GetSubPermissionRules() []shireikan.SubPermission {
	return nil
}

func (c *Help) IsExecutableInDMChannels() bool {
	return true
}

func (c *Help) Exec(ctx shireikan.Context) error {
	emb := utils.GenerateEmbed(ctx, discordgo.MessageEmbed{
		Fields: make([]*discordgo.MessageEmbedField, 0),
	})

	handler, _ := ctx.GetObject(shireikan.ObjectMapKeyHandler).(shireikan.Handler)

	if len(ctx.GetArgs()) == 0 {
		cmds := make(map[string][]shireikan.Command)
		for _, c := range handler.GetCommandInstances() {
			group := c.GetGroup()
			if _, ok := cmds[group]; !ok {
				cmds[group] = make([]shireikan.Command, 0)
			}
			cmds[group] = append(cmds[group], c)
		}

		emb.Title = "Command List"

		for cat, catCmds := range cmds {
			commandHelpLines := ""
			for _, c := range catCmds {
				commandHelpLines += fmt.Sprintf("`%s` - *%s*\n", c.GetInvokes()[0], c.GetDescription())
			}
			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:  cat,
				Value: commandHelpLines,
			})
		}
	} else {
		invoke := ctx.GetArgs().Get(0).AsString()
		cmd, ok := handler.GetCommand(invoke)
		if !ok {
			_, err := ctx.ReplyEmbedError(
				fmt.Sprintf("No command was found with the invoke `%s`.", invoke), "Error")
			return err
		}

		emb.Title = "Command Description"

		description := cmd.GetDescription()
		if description == "" {
			description = "`no description`"
		}

		help := cmd.GetHelp()
		if help == "" {
			help = "`no uage information`"
		}

		emb.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Invokes",
				Value:  strings.Join(cmd.GetInvokes(), "\n"),
				Inline: true,
			},
			{
				Name:   "Group",
				Value:  cmd.GetGroup(),
				Inline: true,
			},
			{
				Name:   "Domain Name",
				Value:  "`" + cmd.GetDomainName() + "`",
				Inline: true,
			},
			{
				Name:   "DM Capable",
				Value:  strconv.FormatBool(cmd.IsExecutableInDMChannels()),
				Inline: true,
			},
			{
				Name:  "Description",
				Value: description,
			},
			{
				Name:  "Usage",
				Value: help,
			},
		}

		if spr := cmd.GetSubPermissionRules(); spr != nil {
			txt := "*`[E]` in front of permissions means `Explicit`, which means that this " +
				"permission must be explicitly allowed and can not be wild-carded.\n" +
				"`[D]` implies that wildecards will apply to this sub permission.*\n\n"

			for _, rule := range spr {
				expl := "D"
				if rule.Explicit {
					expl = "E"
				}

				txt = fmt.Sprintf("%s`[%s]` %s - *%s*\n",
					txt, expl, getTermAssembly(cmd, rule.Term), rule.Description)
			}

			emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
				Name:  "Sub Permission Rules",
				Value: txt,
			})
		}
	}

	_, err := utils.ReplyWithEmbed(ctx, emb)
	if err != nil {
		utils.ErrorHandler(err)
		return err
	}

	return err
}

func getTermAssembly(cmd shireikan.Command, term string) string {
	if strings.HasPrefix(term, "/") {
		return term[1:]
	}
	return cmd.GetDomainName() + "." + term
}
