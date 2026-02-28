package semantics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyp-api-gateway/src/config"
	database "fyp-api-gateway/src/db"
	"log/slog"
	"net/http"

	"gopkg.in/yaml.v3"
)

type AnalysisResult struct {
	Findings []string `json:"findings"`
}

type RouteView struct {
	Host      string
	Port      int
	Path      string
	Auth      bool
	Upstream  string
	RateLimit config.RateLimit
}

type FinalFindings struct {
	Errors  []string `json:"errors"`
	Updates []string `json:"updates"`
}

func RecvConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received config from management plane")
	var newCfg config.GatewayConfig

	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("error getting session cookie ", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	username := database.RetrieveUserBySessionId(cookie.Value)

	if err = yaml.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		slog.Error("Error decoding config", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Analyse the new config file
	slog.Info("analysing new config file...")
	findings, err := analyse(username, newCfg)
	if err != nil {
		slog.Error("Error analysing new config", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// send the findings back to the frontend
	slog.Info("Sending data back to management plane")
	_, err = http.Post("http://management-plane:80/file/findings", "application/json", bytes.NewBuffer(findings))
	if err != nil {
		slog.Error("Error posting new config", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func analyse(username string, newCfg config.GatewayConfig) ([]byte, error) {
	oldCfgStr := database.RetrieveUserConfig(username)

	var oldCfg config.GatewayConfig
	err := yaml.Unmarshal([]byte(oldCfgStr), &oldCfg)
	if err != nil {
		slog.Error("Error decoding config", "error", err)
		return nil, err
	}

	oldConfig := flattenConfig(oldCfg)
	newConfig := flattenConfig(newCfg)
	foundErrors := validateConfigErrors(oldConfig, newConfig)
	foundUpdates := explainDifferences(oldConfig, newConfig)

	finalFindings := FinalFindings{}
	finalFindings.Errors = foundErrors
	finalFindings.Updates = foundUpdates
	findings, err := json.Marshal(finalFindings)
	if err != nil {
		slog.Error("Error marshalling findings map", "error", err)
		return nil, err
	}

	return findings, nil
}

func validateConfigErrors(oldConf, newConf []RouteView) []string {
	var findings []string

	// duplicate routes
	paths := make(map[string]bool)
	for _, r := range newConf {
		if paths[r.Path] {
			findings = append(findings, "Duplicate routes detected: "+r.Path)
		}
		paths[r.Path] = true
	}

	// route shadowing
	for i := 0; i < len(newConf); i++ {
		for j := 0; j < len(oldConf); j++ {
			if i == j {
				continue
			}

			r1 := newConf[i]
			r2 := newConf[j]

			if r1.Host == r2.Host &&
				r1.Port == r2.Port &&
				pathShadows(r1.Path, r2.Path) {
				findings = append(findings,
					"Route "+r2.Path+" may be shadowed by "+r1.Path+" on "+
						r1.Host+":"+fmt.Sprint(r1.Port))
			}
		}
	}

	return findings
}

func explainDifferences(oldConf, newConf []RouteView) []string {
	var findings []string

	oldRoutes := indexRoutes(oldConf)
	newRoutes := indexRoutes(newConf)

	// Detect Added Routes
	for key, newRoute := range newRoutes {
		_, exists := oldRoutes[key]
		if !exists {
			findings = append(findings,
				"New route added: "+newRoute.Path+
					" on "+newRoute.Host)

			if !newRoute.Auth {
				findings = append(findings,
					"New public endpoint exposed at "+newRoute.Path)
			}
		}
	}

	// Detect Removed Routes
	for key, oldRoute := range oldRoutes {
		_, exists := newRoutes[key]
		if !exists {
			findings = append(findings,
				"Route removed: "+oldRoute.Path+
					" on "+oldRoute.Host)

			if oldRoute.Auth {
				findings = append(findings,
					"Previously protected route "+oldRoute.Path+
						" has been removed")
			}
		}
	}

	// Detect Modified Routes
	for key, newRoute := range newRoutes {
		oldRoute, exists := oldRoutes[key]
		if !exists {
			continue
		}

		// Auth Widening
		if oldRoute.Auth && !newRoute.Auth {
			findings = append(findings,
				"Authentication removed from route "+newRoute.Path+
					" on "+newRoute.Host)
		}

		// Auth Tightening
		if !oldRoute.Auth && newRoute.Auth {
			findings = append(findings,
				"Authentication now required for route "+newRoute.Path)
		}

		// Upstream Change
		if oldRoute.Upstream != newRoute.Upstream {
			findings = append(findings,
				"Traffic for "+newRoute.Path+" will be routed from "+
					oldRoute.Upstream+" to "+newRoute.Upstream)
		}

		// Rate Limit Tightening
		if newRoute.RateLimit.Rate < oldRoute.RateLimit.Rate {
			findings = append(findings,
				"Rate limit tightened on "+newRoute.Path+
					" ("+fmt.Sprint(oldRoute.RateLimit.Rate)+
					" → "+fmt.Sprint(newRoute.RateLimit.Rate)+")")
		}

		// Rate Limit Relaxed
		if newRoute.RateLimit.Rate > oldRoute.RateLimit.Rate {
			findings = append(findings,
				"Rate limit relaxed on "+newRoute.Path+
					" ("+fmt.Sprint(oldRoute.RateLimit.Rate)+
					" → "+fmt.Sprint(newRoute.RateLimit.Rate)+")")
		}
	}

	return findings
}

func flattenConfig(cfg config.GatewayConfig) []RouteView {
	var routes []RouteView

	for _, c := range cfg.Connections {
		for _, r := range c.Routes {
			routes = append(routes, RouteView{
				Host:      c.Host,
				Port:      c.Port,
				Path:      r.Path,
				Auth:      r.Auth,
				Upstream:  r.Upstream.Name,
				RateLimit: r.RateLimit,
			})
		}
	}
	return routes
}

func indexRoutes(routes []RouteView) map[string]RouteView {
	index := make(map[string]RouteView)
	for _, r := range routes {
		key := r.Host + ":" + fmt.Sprint(r.Port) + r.Path
		index[key] = r
	}
	return index
}

func pathShadows(a, b string) bool {
	if a == "/" {
		return true
	}
	return len(b) > len(a) && b[:len(a)] == a
}
