package integrations

import (
	"fmt"

	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/dgc"
)

func GetWorkspaceIntegrationForCommand(ctx *dgc.Ctx, integrationID string) (workspaceIntegrationID int, err error) {
	data := ctx.CustomObjects.MustGet("workspaceIntegrations")

	workspaceIntegrations := data.([]types.WorkspaceIntegration)

	for _, workspaceIntegration := range workspaceIntegrations {
		if workspaceIntegration.Integration == integrationID {
			return workspaceIntegration.ID, nil
		}
	}

	return 0, fmt.Errorf("no workspace integration found for integration %s", integrationID)
}
