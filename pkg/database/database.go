package database

import (
	"github.com/astralservices/bots/pkg/types"
)

type Database interface {
	////////////////////////////////////////////////////////////////////////////
	//// GLOBAL SETTINGS ///////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	GetAllBotsForRegion(region string) ([]types.Bot, error)
	GetRegion(region string) (types.Region, error)

	////////////////////////////////////////////////////////////////////////////
	//// BOT SETTINGS //////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	// GetBot returns the bot settings for the given bot.
	GetBot(botID string) (types.Bot, error)
	// SetBot sets the bot settings for the given bot.
	SetBot(botID string, settings types.Bot) error

	////////////////////////////////////////////////////////////////////////////
	//// BOT PERMISSIONS ///////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////////
	//// REPORTS ///////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	// AddReport adds a report to the database.
	AddReport(report types.Report) (types.Report, error)
	// DeleteReport deletes a report from the database.
	DeleteReport(reportID string) error
	// GetReport returns a report from the database.
	GetReport(reportID string) (types.Report, error)
	// GetReports returns all reports from the database for the given guild.
	GetReports(guildID string) ([]types.Report, error)
	// GetReportsFiltered returns all reports from the database for the given guild,
	// filtered by the given filter.
	GetReportsFiltered(guildID string, filter types.ReportFilter) ([]types.Report, error)
	// UpdateReport updates a report in the database.
	UpdateReport(report types.Report) error

	////////////////////////////////////////////////////////////////////////////
	//// PROVIDERS /////////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	// GetProviderForUser returns the provider for the given user.
	GetProviderForUser(userID string, providerID string) (types.Provider, error)

	// GetProviderFromDiscord returns the provider for the given discord user.
	GetProviderFromDiscord(userID string, providerID string) (types.Provider, error)

	////////////////////////////////////////////////////////////////////////////
	//// INTEGRATIONS //////////////////////////////////////////////////////////
	////////////////////////////////////////////////////////////////////////////

	// GetIntegrationDataForUser returns the integration data for the given user.
	GetIntegrationDataForUser(userID string, integrationID string, workspaceIntegrationID int) (types.IntegrationData, error)

	// GetIntegrationDataForWorkspace returns the integration data for the given workspace.
	GetIntegrationDataForWorkspace(workspaceID string, integrationID string) ([]types.IntegrationData, error)

	// GetIntegrationForWorkspace returns the integration for the given workspace.
	GetIntegrationForWorkspace(integrationID string, workspaceID string) (types.WorkspaceIntegration, error)

	// GetIntegrationsForWorkspace returns the integrations for the given workspace.
	GetIntegrationsForWorkspace(workspaceID string) ([]types.WorkspaceIntegration, error)
}
