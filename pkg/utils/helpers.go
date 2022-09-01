package utils

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/astralservices/bots/pkg/constants"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

const (
	CategoryBot        = "Bot"
	CategoryFun        = "Fun"
	CategoryImage      = "Image"
	CategoryMusic      = "Music"
	CategoryModeration = "Moderation"
	CategoryOwner      = "Owner"
	CategoryUtils      = "Utils"
	CategoryVoice      = "Voice"
	CategoryWeb        = "Web"
)

func GenerateEmbed(ctx dgc.Ctx, e discordgo.MessageEmbed) *discordgo.MessageEmbed {
	embed := discordgo.MessageEmbed{
		Title:       e.Title,
		Description: e.Description,
		Fields:      e.Fields,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    ctx.Event.Author.Username,
			IconURL: ctx.Event.Author.AvatarURL("48"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Astral Bots " + constants.VERSION,
			IconURL: ctx.Session.State.User.AvatarURL("48"),
		},
		Timestamp: ctx.Event.Timestamp.Format(time.RFC3339),
	}

	if e.Color != 0 {
		embed.Color = e.Color
	} else {
		embed.Color = 0x4f4bfe
	}

	if e.Image != nil {
		embed.Image = e.Image
	}

	return &embed
}

func ErrorEmbed(ctx dgc.Ctx, err error) *discordgo.MessageEmbed {
	return GenerateEmbed(ctx, discordgo.MessageEmbed{
		Title:       "Error",
		Description: err.Error(),
		Color:       0xff0000,
	})
}

// IsAdmin returns true if one of the members roles has
// admin (0x8) permissions on the passed guild.
func IsAdmin(g *discordgo.Guild, m *discordgo.Member) bool {
	if m == nil || g == nil {
		return false
	}

	for _, r := range g.Roles {
		if r.Permissions&0x8 != 0 {
			for _, mrID := range m.Roles {
				if r.ID == mrID {
					return true
				}
			}
		}
	}

	return false
}

func GetUserFromAstralId(s *discordgo.Session, id string, db db.SupabaseMiddleware) (*discordgo.User, error) {
	provider, err := db.GetProviderForUser(id, "discord")

	if err != nil {
		return nil, err
	}

	return s.User(provider.ProviderID)
}

func GetAstralIdFromUser(s *discordgo.Session, user string, db db.SupabaseMiddleware) (string, error) {
	provider, err := db.GetProviderFromDiscord(user, "discord")

	if err != nil {
		return "", err
	}

	return *provider.ID, nil
}

func SortRoles(r []*discordgo.Role, reversed bool) {
	var f func(i, j int) bool

	if reversed {
		f = func(i, j int) bool {
			return r[i].Position > r[j].Position
		}
	} else {
		f = func(i, j int) bool {
			return r[i].Position < r[j].Position
		}
	}

	sort.Slice(r, f)
}

func GetSortedMemberRoles(s *discordgo.Session, guildID, memberID string, reversed bool, includeEveryone bool) ([]*discordgo.Role, error) {
	member, err := s.GuildMember(guildID, memberID)
	if err != nil {
		return nil, err
	}

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	rolesMap := make(map[string]*discordgo.Role)
	for _, r := range roles {
		rolesMap[r.ID] = r
	}

	membRoles := make([]*discordgo.Role, len(member.Roles)+1)
	applied := 0
	for _, rID := range member.Roles {
		if r, ok := rolesMap[rID]; ok {
			membRoles[applied] = r
			applied++
		}
	}

	if includeEveryone {
		membRoles[applied] = rolesMap[guildID]
		applied++
	}

	membRoles = membRoles[:applied]

	SortRoles(membRoles, reversed)

	return membRoles, nil
}

var (
	RoleCheckFuncs = []func(*discordgo.Role, string) bool{
		// 1. ID exact match
		func(r *discordgo.Role, resolvable string) bool {
			return r.ID == resolvable
		},
		// 2. name exact match
		func(r *discordgo.Role, resolvable string) bool {
			return r.Name == resolvable
		},
		// 3. name lowercased exact match
		func(r *discordgo.Role, resolvable string) bool {
			return strings.ToLower(r.Name) == strings.ToLower(resolvable)
		},
		// 4. name lowercased startswith
		func(r *discordgo.Role, resolvable string) bool {
			return strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(resolvable))
		},
		// 5. name lowercased contains
		func(r *discordgo.Role, resolvable string) bool {
			return strings.Contains(strings.ToLower(r.Name), strings.ToLower(resolvable))
		},
	}

	MemberCheckFuncs = []func(*discordgo.Member, string) bool{
		// 1. ID exact match
		func(r *discordgo.Member, resolvable string) bool {
			return r.User.ID == resolvable
		},
		// 2. username exact match
		func(r *discordgo.Member, resolvable string) bool {
			return r.User.Username == resolvable
		},
		// 3. username lowercased exact match
		func(r *discordgo.Member, resolvable string) bool {
			return strings.ToLower(r.User.Username) == strings.ToLower(resolvable)
		},
		// 4. username lowercased startswith
		func(r *discordgo.Member, resolvable string) bool {
			return strings.HasPrefix(strings.ToLower(r.User.Username), strings.ToLower(resolvable))
		},
		// 5. username lowercased contains
		func(r *discordgo.Member, resolvable string) bool {
			return strings.Contains(strings.ToLower(r.User.Username), strings.ToLower(resolvable))
		},
		// 6. nick exact match
		func(r *discordgo.Member, resolvable string) bool {
			return r.Nick == resolvable
		},
		// 7. nick lowercased exact match
		func(r *discordgo.Member, resolvable string) bool {
			return r.Nick != "" && strings.ToLower(r.Nick) == strings.ToLower(resolvable)
		},
		// 8. nick lowercased starts with
		func(r *discordgo.Member, resolvable string) bool {
			return r.Nick != "" && strings.HasPrefix(strings.ToLower(r.Nick), strings.ToLower(resolvable))
		},
		// 9. nick lowercased contains
		func(r *discordgo.Member, resolvable string) bool {
			return r.Nick != "" && strings.Contains(strings.ToLower(r.Nick), strings.ToLower(resolvable))
		},
	}

	ChannelCheckFuncs = []func(*discordgo.Channel, string) bool{
		// 1. ID exact match
		func(r *discordgo.Channel, resolvable string) bool {
			return r.ID == resolvable
		},
		// 2. mention exact match
		func(r *discordgo.Channel, resolvable string) bool {
			l := len(resolvable)
			return l > 3 && r.ID == resolvable[2:l-1]
		},
		// 3. name exact match
		func(r *discordgo.Channel, resolvable string) bool {
			return r.Name == resolvable
		},
		// 4. name lowercased exact match
		func(r *discordgo.Channel, resolvable string) bool {
			return strings.ToLower(r.Name) == strings.ToLower(resolvable)
		},
		// 5. name lowercased starts with
		func(r *discordgo.Channel, resolvable string) bool {
			return strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(resolvable))
		},
		// 6. name lowercased contains
		func(r *discordgo.Channel, resolvable string) bool {
			return strings.Contains(strings.ToLower(r.Name), strings.ToLower(resolvable))
		},
	}
)

type DataOutlet interface {
	GuildRoles(guildID string) ([]*discordgo.Role, error)
	GuildMembers(guildID string, after string, limit int) (st []*discordgo.Member, err error)
	GuildChannels(guildID string) (st []*discordgo.Channel, err error)
}

// FetchMember tries to fetch a member on the specified guild
// by given resolvable and returns this member, when found.
// You can pass a condition function which ignores the result
// if this functions returns false on the given object.
// If no object was found, ErrNotFound is returned.
// If any other unexpected error occurs during fetching,
// this error is returned as well.
func FetchMember(s DataOutlet, guildID, resolvable string, condition ...func(*discordgo.Member) bool) (*discordgo.Member, error) {
	rx := regexp.MustCompile("<@|!|>")
	resolvable = rx.ReplaceAllString(resolvable, "")
	var lastUserID string

	for {
		members, err := s.GuildMembers(guildID, lastUserID, 1000)
		if err != nil {
			return nil, err
		}

		if len(members) < 1 {
			break
		}

		lastUserID = members[len(members)-1].User.ID

		for _, checkFunc := range MemberCheckFuncs {
			for _, m := range members {
				if len(condition) > 0 && condition[0] != nil {
					if !condition[0](m) {
						continue
					}
				}
				if checkFunc(m, resolvable) {
					return m, nil
				}
			}
		}
	}

	return nil, errors.New("not found")
}
