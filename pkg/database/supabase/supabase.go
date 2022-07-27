package db

import (
	"os"

	"github.com/astralservices/bots/pkg/database"
	"github.com/astralservices/bots/pkg/types"
	"github.com/nedpals/supabase-go"
)

type SupabaseMiddleware struct {
	Supabase *supabase.Client
}

var _ database.Database = (*SupabaseMiddleware)(nil)

func New() SupabaseMiddleware {
	supabaseClient := supabase.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"))
	return SupabaseMiddleware{Supabase: supabaseClient}
}

func (s SupabaseMiddleware) GetAllBotsForRegion(region string) ([]types.Bot, error) {
	var bots []types.Bot
	err := s.Supabase.DB.From("bots").Select("*").Eq("region", region).Execute(&bots)
	return bots, err
}

func (s SupabaseMiddleware) GetRegion(region string) (types.Region, error) {
	var regionData types.Region
	err := s.Supabase.DB.From("regions").Select("*").Single().Eq("id", region).Execute(&regionData)
	return regionData, err
}

func (m *SupabaseMiddleware) GetBot(botID string) (types.Bot, error) {
	var bot types.Bot
	err := m.Supabase.DB.From("bots").Select("*").Single().Eq("id", botID).Execute(&bot)
	return bot, err
}

func (m *SupabaseMiddleware) SetBot(botID string, settings types.Bot) error {
	return m.Supabase.DB.From("bots").Update(settings).Eq("id", botID).Execute(nil)
}

func (m *SupabaseMiddleware) AddReport(report types.Report) error {
	return m.Supabase.DB.From("moderation_actions").Insert(report).Execute(nil)
}

func (m *SupabaseMiddleware) DeleteReport(reportID string) error {
	return m.Supabase.DB.From("moderation_actions").Delete().Eq("id", reportID).Execute(nil)
}

func (m *SupabaseMiddleware) GetReport(reportID string) (types.Report, error) {
	var report types.Report
	err := m.Supabase.DB.From("moderation_actions").Select("*").Single().Eq("id", reportID).Execute(&report)
	return report, err
}

func (m *SupabaseMiddleware) GetReports(guildID string) ([]types.Report, error) {
	var reports []types.Report
	err := m.Supabase.DB.From("moderation_actions").Select("*").Eq("guild", guildID).Execute(&reports)
	return reports, err
}

func (m *SupabaseMiddleware) GetReportsFiltered(guildID string, filter types.ReportFilter) ([]types.Report, error) {
	var reports []types.Report
	sel := m.Supabase.DB.From("moderation_actions").Select("*")

	if filter.Page > 0 && filter.Size > 0 {
		sel = sel.LimitWithOffset(filter.Size, filter.Page*filter.Size)
	}

	query := sel.Eq("guild", guildID)

	if filter.Action != "" {
		query = query.Eq("action", filter.Action)
	}
	if filter.Moderator != "" {
		query = query.Eq("moderator", filter.Moderator)
	}
	if filter.User != "" {
		query = query.Eq("user", filter.User)
	}
	if filter.Expired {
		query = query.Eq("expired", "true")
	}

	err := query.Execute(&reports)

	return reports, err
}

func (m *SupabaseMiddleware) UpdateReport(report types.Report) error {
	return m.Supabase.DB.From("moderation_actions").Update(report).Eq("id", *report.ID).Execute(nil)
}

func (m *SupabaseMiddleware) GetProviderForUser(userID string, providerID string) (types.Provider, error) {
	var provider types.Provider
	err := m.Supabase.DB.From("providers").Select("*").Single().Eq("user", userID).Eq("id", providerID).Execute(&provider)
	return provider, err
}
