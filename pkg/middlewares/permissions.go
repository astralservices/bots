package middlewares

import (
	"fmt"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/permissions"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

// PermissionsMiddleware is a command handler middleware
// processing permissions for command execution.
//
// Implements the shireikan.Middleware interface and
// exposes functions to check permissions.
type PermissionsMiddleware struct {
	Bot types.Bot

	db    *db.SupabaseMiddleware
	cfg   *types.Bot
	cache map[string]permissions.PermissionArray
}

func (m *PermissionsMiddleware) UpdateConfig(cfg *types.Bot) {
	m.cfg = cfg
	// force update the cache
	m.cache = make(map[string]permissions.PermissionArray)
}

func (m *PermissionsMiddleware) Handle(next dgc.ExecutionHandler) dgc.ExecutionHandler {
	return func(ctx *dgc.Ctx) {
		if m.db == nil {
			db := db.New()
			m.db = &db
		}

		if m.cfg == nil {
			m.cfg = &m.Bot
		}

		guildID := ctx.Event.Message.GuildID

		ok, _, err := m.CheckPermissions(ctx.Session, guildID, ctx.Message.Author.ID, ctx.Command.Domain)

		if err != nil {
			err := ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("Error while checking permissions\nError: %s", err.Error())))
			if err != nil {
				utils.ErrorHandler(err)
			}
			return
		}

		if !ok {
			err := ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("You do not have the permission to execute this command")))
			if err != nil {
				utils.ErrorHandler(err)
			}
			return
		}

		next(ctx)
	}
}

// GetPermissions tries to fetch the permissions array of
// the passed user of the specified guild. The merged
// permissions array is returned as well as the override,
// which is true when the specified user is the bot owner,
// guild owner or an admin of the guild.
func (m *PermissionsMiddleware) GetPermissions(s *discordgo.Session, guildID, userID string) (perm permissions.PermissionArray, overrideExplicits bool, err error) {
	if m.cache == nil {
		m.cache = make(map[string]permissions.PermissionArray)
	}

	if p, ok := m.cache[userID]; ok {
		return p, false, nil
	}

	if guildID != "" {
		perm, err = m.GetMemberPermission(s, guildID, userID)
		if err != nil {
			return
		}
	} else {
		perm = make(permissions.PermissionArray, 0)
	}

	discordUser, err := utils.GetUserFromAstralId(s, *m.cfg.Owner, *m.db)

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

			hardCoded := []string{
				"+astral.integrations.reactionrole-add",
				"+astral.integrations.reactionrole-remove",
			}

			perm = perm.Merge(defAdminRoles, false)
			perm = perm.Merge(hardCoded, false)
			overrideExplicits = true
		}
	}

	var defUserRoles []string
	defUserRoles = m.cfg.Permissions.DefaultUserRules

	perm = perm.Merge(defUserRoles, false)

	var hardCoded []string
	hardCoded = []string{
		"-astral.integrations.reactionrole-add",
		"-astral.integrations.reactionrole-remove",
		"+astral.integrations.reactionrole-list",
	}

	perm = perm.Merge(hardCoded, false)

	m.cache[userID] = perm

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
	guildPerms := m.cfg.Permissions.Roles
	userPerms := m.cfg.Permissions.Users

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

	if p, ok := userPerms[memberID]; ok {
		if res == nil {
			res = p
		} else {
			res = res.Merge(p, true)
		}
	}

	return res, nil
}
