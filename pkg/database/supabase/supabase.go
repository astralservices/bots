package db

import (
	"os"

	"github.com/astralservices/bots/pkg/database"
	"github.com/astralservices/bots/pkg/types"
	"github.com/grid-rbx/supabase-go"
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
	_, err := s.Supabase.DB.From("bots").Select("*", "", false).Eq("region", region).ExecuteTo(&bots)
	return bots, err
}

func (s SupabaseMiddleware) GetRegion(region string) (types.Region, error) {
	var regionData types.Region
	_, err := s.Supabase.DB.From("regions").Select("*", "", false).Single().Eq("id", region).ExecuteTo(&regionData)
	return regionData, err
}

func (m *SupabaseMiddleware) GetBot(botID string) (types.Bot, error) {
	var bot types.Bot
	_, err := m.Supabase.DB.From("bots").Select("*", "", false).Single().Eq("id", botID).ExecuteTo(&bot)
	return bot, err
}

func (m *SupabaseMiddleware) SetBot(botID string, settings types.Bot) error {
	_, err := m.Supabase.DB.From("bots").Update(settings, "", "").Eq("id", botID).ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) AddReport(report types.Report) error {
	_, err := m.Supabase.DB.From("moderation_actions").Insert(report, false, "", "", "").ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) DeleteReport(reportID string) error {
	_, err := m.Supabase.DB.From("moderation_actions").Delete("", "").Eq("id", reportID).ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) GetReport(reportID string) (types.Report, error) {
	var report types.Report
	_, err := m.Supabase.DB.From("moderation_actions").Select("*", "", false).Single().Eq("id", reportID).ExecuteTo(&report)
	return report, err
}

func (m *SupabaseMiddleware) GetReports(guildID string) ([]types.Report, error) {
	var reports []types.Report
	_, err := m.Supabase.DB.From("moderation_actions").Select("*", "", false).Eq("guild", guildID).ExecuteTo(&reports)
	return reports, err
}

func (m *SupabaseMiddleware) GetReportsFiltered(guildID string, filter types.ReportFilter) ([]types.Report, error) {
	var reports []types.Report
	sel := m.Supabase.DB.From("moderation_actions").Select("*", "", false)
	if filter.Page > 0 && filter.Size > 0 {
		from, to := database.GetPagination(filter.Page, filter.Size)
		sel = sel.Range(from, to, "")
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

	_, err := query.ExecuteTo(&reports)

	return reports, err
}

func (m *SupabaseMiddleware) UpdateReport(report types.Report) error {
	_, err := m.Supabase.DB.From("moderation_actions").Update(report, "", "").Eq("id", *report.ID).ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) GetProviderForUser(userID string, providerID string) (types.Provider, error) {
	var provider types.Provider
	_, err := m.Supabase.DB.From("providers").Select("*", "", false).Single().Eq("id", userID).Eq("type", providerID).ExecuteTo(&provider)
	return provider, err
}
