// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/core/settings"
	"github.com/DataDog/datadog-agent/comp/core/status"
	"github.com/DataDog/datadog-agent/comp/core/workloadmeta"
	pkgconfig "github.com/DataDog/datadog-agent/pkg/config"
)

//nolint:revive // TODO(PROC) Fix revive linter
type APIServerDeps struct {
	fx.In

	Config       config.Component
	Log          log.Component
	WorkloadMeta workloadmeta.Component
	Status       status.Component
	Settings     settings.Component
}

func injectDeps(deps APIServerDeps, handler func(APIServerDeps, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		handler(deps, writer, req)
	}
}

//nolint:revive // TODO(PROC) Fix revive linter
func SetupAPIServerHandlers(deps APIServerDeps, r *mux.Router) {
	r.HandleFunc("/config", deps.Settings.GetFullConfig(pkgconfig.Datadog, "process_config")).Methods("GET")
	r.HandleFunc("/config/all", deps.Settings.GetFullConfig(pkgconfig.Datadog, "")).Methods("GET") // Get all fields from process-agent Config object
	r.HandleFunc("/config/list-runtime", deps.Settings.ListConfigurable).Methods("GET")
	r.HandleFunc("/config/{setting}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		setting := vars["setting"]
		deps.Settings.GetValue(setting, w, r)
	}).Methods("GET")
	r.HandleFunc("/config/{setting}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		setting := vars["setting"]
		deps.Settings.SetValue(setting, w, r)
	}).Methods("POST")

	r.HandleFunc("/agent/status", injectDeps(deps, statusHandler)).Methods("GET")
	r.HandleFunc("/agent/tagger-list", injectDeps(deps, getTaggerList)).Methods("GET")
	r.HandleFunc("/agent/workload-list/short", func(w http.ResponseWriter, r *http.Request) {
		workloadList(w, false, deps.WorkloadMeta)
	}).Methods("GET")
	r.HandleFunc("/agent/workload-list/verbose", func(w http.ResponseWriter, r *http.Request) {
		workloadList(w, true, deps.WorkloadMeta)
	}).Methods("GET")
	r.HandleFunc("/check/{check}", checkHandler).Methods("GET")
}
