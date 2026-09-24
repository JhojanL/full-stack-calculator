package httpapi

import "net/http"

// healthcheck reports process availability, environment, and build version to w.
// Request r supplies logging context; no downstream dependencies are probed.
func (app *api) healthcheck(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, r, http.StatusOK, envelope{
		"status":      "available",
		"system_info": map[string]string{"environment": app.config.Environment, "version": app.version},
	})
}
