package db

import (
	"os"
	"strconv"
	"time"

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

func (m *SupabaseMiddleware) AddReport(report types.Report) (types.Report, error) {
	var newReport types.Report
	_, err := m.Supabase.DB.From("moderation_actions").Insert(report, false, "", "", "").Single().ExecuteTo(&newReport)
	return newReport, err
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

func (m *SupabaseMiddleware) GetReportsFiltered(filter types.ReportFilter) ([]types.Report, error) {
	var reports []types.Report
	sel := m.Supabase.DB.From("moderation_actions").Select("*", "", false)
	if filter.Page > 0 && filter.Size > 0 {
		from, to := database.GetPagination(filter.Page, filter.Size)
		sel = sel.Range(from, to, "")
	}

	query := sel

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
		// if the expiry date is in the past, it's expired
		query = query.Lte("expiry", time.Now().Format(time.RFC3339)).Eq("expired", "false")
	}
	if filter.Guild != "" {
		query = query.Eq("guild", filter.Guild)
	}
	if filter.Bot != "" {
		query = query.Eq("bot", filter.Bot)
	}

	_, err := query.ExecuteTo(&reports)

	return reports, err
}

func (m *SupabaseMiddleware) UpdateReport(report types.Report) error {
	_, err := m.Supabase.DB.From("moderation_actions").Update(report, "", "").Eq("id", *report.ID).ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) ExpireReport(reportID string) error {
	report, err := m.GetReport(reportID)
	if err != nil {
		return err
	}

	report.Expired = true

	return m.UpdateReport(report)
}

func (m *SupabaseMiddleware) GetProviderForUser(userID string, providerID string) (types.Provider, error) {
	var provider types.Provider
	_, err := m.Supabase.DB.From("providers").Select("*", "", false).Single().Eq("id", userID).Eq("type", providerID).ExecuteTo(&provider)
	return provider, err
}

func (m *SupabaseMiddleware) GetProviderFromDiscord(userID string, providerID string) (types.Provider, error) {
	var discord, provider types.Provider
	_, err := m.Supabase.DB.From("providers").Select("*", "", false).Single().Eq("provider_id", userID).Eq("type", "discord").ExecuteTo(&discord)

	if err != nil {
		return types.Provider{}, err
	}

	if providerID == "discord" {
		return discord, nil
	}

	_, err = m.Supabase.DB.From("providers").Select("*", "", false).Single().Eq("user", *discord.ID).Eq("type", providerID).ExecuteTo(&provider)

	return provider, err
}

func (m *SupabaseMiddleware) GetIntegrationDataForUser(userID string, integrationID string, workspaceIntegrationID int) (types.IntegrationData, error) {
	var integrationData types.IntegrationData
	_, err := m.Supabase.DB.From("integration_data").Select("*", "", false).Single().Eq("user", userID).Eq("integration", integrationID).Eq("workspaceIntegration", strconv.Itoa(workspaceIntegrationID)).ExecuteTo(&integrationData)
	return integrationData, err
}

func (m *SupabaseMiddleware) SetIntegrationDataForUser(userID string, integrationID string, workspaceIntegrationID int, data any) error {
	i, err := m.GetIntegrationDataForUser(userID, integrationID, workspaceIntegrationID)
	if err != nil {
		return err
	}

	i.Data = data

	_, err = m.Supabase.DB.From("integration_data").Update(i, "", "").Eq("user", userID).Eq("integration", integrationID).Eq("workspaceIntegration", strconv.Itoa(workspaceIntegrationID)).ExecuteTo(nil)
	return err
}

func (m *SupabaseMiddleware) GetIntegrationDataForWorkspace(workspaceID string, integrationID string) ([]types.IntegrationData, error) {
	i, err := m.GetIntegrationForWorkspace(integrationID, workspaceID)

	if err != nil {
		return nil, err
	}

	var integrationData []types.IntegrationData
	_, err = m.Supabase.DB.From("integration_data").Select("*", "", false).Eq("workspaceIntegration", strconv.Itoa(i.ID)).Eq("integration", integrationID).ExecuteTo(&integrationData)
	return integrationData, err
}

func (m *SupabaseMiddleware) GetIntegrationForWorkspace(integrationID string, workspaceID string) (types.WorkspaceIntegration, error) {
	var integration types.WorkspaceIntegration
	_, err := m.Supabase.DB.From("workspace_integrations").Select("*", "", false).Single().Eq("integration", integrationID).Eq("workspace", workspaceID).ExecuteTo(&integration)
	return integration, err
}

func (m *SupabaseMiddleware) GetIntegrationsForWorkspace(workspaceID string) ([]types.WorkspaceIntegration, error) {
	var integrations []types.WorkspaceIntegration
	_, err := m.Supabase.DB.From("workspace_integrations").Select("*", "", false).Eq("workspace", workspaceID).ExecuteTo(&integrations)
	return integrations, err
}

func (m *SupabaseMiddleware) GetStatistics(botID string) ([]types.BotAnalytics, error) {
	var statistics []types.BotAnalytics
	_, err := m.Supabase.DB.From("bot_analytics").Select("*", "", false).Eq("bot", botID).ExecuteTo(&statistics)
	return statistics, err
}
