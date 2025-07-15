package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/trustyai-explainability/trustyai-dashboard/bff/internal/constants"
	"github.com/trustyai-explainability/trustyai-dashboard/bff/internal/integrations/kubernetes"
	"github.com/trustyai-explainability/trustyai-dashboard/bff/internal/models"
)

type UserEnvelope Envelope[*models.User, None]

func (app *App) UserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ctx := r.Context()
	identity, ok := ctx.Value(constants.RequestIdentityKey).(*kubernetes.RequestIdentity)
	if !ok || identity == nil {
		app.badRequestResponse(w, r, fmt.Errorf("missing RequestIdentity in context"))
		return
	}

	client, err := app.kubernetesClientFactory.GetClient(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("failed to get Kubernetes client: %w", err))
		return
	}

	user, err := client.GetUser(identity)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	isAdmin, err := client.IsClusterAdmin(identity)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userRes := UserEnvelope{
		Data: &models.User{
			UserID:       user,
			ClusterAdmin: isAdmin,
		},
	}

	err = app.WriteJSON(w, http.StatusOK, userRes, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
