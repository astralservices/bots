package reactionroles

import (
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var ReactionRoleCommand = &dgc.Command{
	Name:        "rr",
	Aliases:     []string{"rr", "reactionrole"},
	Description: "Parent command of all reaction role commands.",
	Usage:       "rr <subcommand>",
	Example:     "rr add",
	Category:    "Reaction Roles",
	Domain:      "astral.integrations.reactionroles",
	Handler: func(ctx *dgc.Ctx) {
		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Error",
			Description: "You must provide a subcommand.",
			Color:       0xff0000,
		}))
	},
	SubCommands: []*dgc.Command{
		AddReactionRole,
		RemoveReactionRole,
		ListReactionRoles,
	},
}
