package middlewares

import (
	"errors"
	"fmt"
	"log"

	"github.com/astralservices/bots/packages/permissions"
	db "github.com/astralservices/bots/supabase"
	"github.com/astralservices/bots/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/nedpals/supabase-go"

	"github.com/zekroTJA/shireikan"
)

// PermissionsMiddleware is a command handler middleware
// processing permissions for command execution.
//
// Implements the shireikan.Middleware interface and
// exposes functions to check permissions.
type PermissionsMiddleware struct {
	db  *supabase.Client
	cfg *utils.IBot
}

func (m *PermissionsMiddleware) Handle(cmd shireikan.Command, ctx shireikan.Context, layer shireikan.MiddlewareLayer) (next bool, err error) {
	if m.db == nil {
		m.db = db.New()
	}

	if m.cfg == nil {
		cfg := ctx.GetObject("bot").(utils.IBot)

		m.cfg = &cfg
	}

	var guildID string
	if ctx.GetGuild() != nil {
		guildID = ctx.GetGuild().ID
	}

	ok, _, err := m.CheckPermissions(ctx.GetSession(), guildID, ctx.GetUser().ID, cmd.GetDomainName())

	if err != nil {
		return false, err
	}

	if !ok {
		utils.ReplyWithEmbed(ctx, utils.ErrorEmbed(ctx, errors.New("You are not allowed to execute this command.")))
		return false, nil
	}

	return true, nil
}

func (m *PermissionsMiddleware) GetLayer() shireikan.MiddlewareLayer {
	return shireikan.LayerBeforeCommand
}

// GetPermissions tries to fetch the permissions array of
// the passed user of the specified guild. The merged
// permissions array is returned as well as the override,
// which is true when the specified user is the bot owner,
// guild owner or an admin of the guild.
func (m *PermissionsMiddleware) GetPermissions(s *discordgo.Session, guildID, userID string) (perm permissions.PermissionArray, overrideExplicits bool, err error) {
	if guildID != "" {
		perm, err = m.GetMemberPermission(s, guildID, userID)
		if err != nil {
			return
		}
	} else {
		perm = make(permissions.PermissionArray, 0)
	}

	discordUser, err := utils.GetUserFromAstralId(s, *m.cfg.Owner, m.db)

	log.Println(discordUser.ID, userID)

	if err != nil {
		return nil, false, err
	}

	if discordUser.ID == userID {
		perm = perm.Merge(permissions.PermissionArray{"+astral.*"}, false)
		overrideExplicits = true
	}

	if guildID != "" {
		guild, err := s.Guild(guildID)
		if err != nil {
			return permissions.PermissionArray{}, false, nil
		}

		member, _ := s.GuildMember(guildID, userID)

		if userID == guild.OwnerID || (member != nil && utils.IsAdmin(guild, member)) {
			var defAdminRoles []string
			defAdminRoles = m.cfg.Permissions.DefaultAdminRules

			perm = perm.Merge(defAdminRoles, false)
			overrideExplicits = true
		}
	}

	var defUserRoles []string
	defUserRoles = m.cfg.Permissions.DefaultUserRules

	perm = perm.Merge(defUserRoles, false)

	fmt.Printf("%+v\n", perm)

	return perm, overrideExplicits, nil
}

// CheckPermissions tries to fetch the permissions of the specified user
// on the specified guild and returns true, if the passed dn matches the
// fetched permissions array. Also, the override status is returned as
// well as errors occured during permissions fetching.
func (m *PermissionsMiddleware) CheckPermissions(s *discordgo.Session, guildID, userID, dn string) (bool, bool, error) {
	perms, overrideExplicits, err := m.GetPermissions(s, guildID, userID)
	if err != nil {
		return false, false, err
	}

	return perms.Check(dn), overrideExplicits, nil
}

// GetMemberPermissions returns a PermissionsArray based on the passed
// members roles permissions rulesets for the given guild.
func (m *PermissionsMiddleware) GetMemberPermission(s *discordgo.Session, guildID string, memberID string) (permissions.PermissionArray, error) {
	guildPerms := m.cfg.Permissions.Users

	membRoles, err := utils.GetSortedMemberRoles(s, guildID, memberID, false, true)
	if err != nil {
		return nil, err
	}

	var res permissions.PermissionArray
	for _, r := range membRoles {
		if p, ok := guildPerms[r.ID]; ok {
			if res == nil {
				res = p
			} else {
				res = res.Merge(p, true)
			}
		}
	}

	return res, nil
}
