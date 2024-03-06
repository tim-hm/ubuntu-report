package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"

	"github.com/gorilla/mux"
	"github.com/ubuntu/ubuntu-report/internal/metrics"
	"golang.org/x/exp/slog"
)

// httpLogger logs the incoming request
func httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info(fmt.Sprintf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// healthCheckHandler handles the health check requests
func (d *Daemon) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Perform any self-test or health check logic here
	// For simplicity, just sending a confirmation response
	// But ultimately we should verify that we can log records
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Health Check: Service is running")
}

type AboutResponse struct {
	Version string `json:"version"`
}

func (d *Daemon) daemonAboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")
	response := AboutResponse{Version: "0.0.2"}
	json.NewEncoder(w).Encode(response)
}

func (d *Daemon) daemonMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	report := `# HELP ubuntu_reportd_requests_total The total number of processed requests
# TYPE ubuntu_reportd_requests_total counter
ubuntu_reportd_requests_total 1234

# HELP ubuntu_reportd_errors_total The total number of errors
# TYPE ubuntu_reportd_errors_total counter
ubuntu_reportd_errors_total 56

# HELP ubuntu_reportd_uptime_seconds The number of seconds the service has been up
# TYPE ubuntu_reportd_uptime_seconds gauge
ubuntu_reportd_uptime_seconds 3600
`
	fmt.Fprint(w, report)
}

// submitHandler handles the POST requests
func (d *Daemon) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	reject := false

	// Retrieve distro/variant/version from URL
	urlvars := mux.Vars(r)
	distro := urlvars["distro"]
	variant := urlvars["variant"]
	version := urlvars["version"]

	// Reject unsupported distros and variants

	if !slices.Contains(d.distros, distro) {
		reject = true
	}
	if !slices.Contains(d.variants, variant) {
		reject = true
	}

	// TODO Reject invalid versions
	if !validateVersion(version) {
		reject = true
	}

	// Read the body of the request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Reject invalid json
	var metrics metrics.MetricsData
	if err := json.Unmarshal(body, &metrics); err != nil {
		http.Error(w, "Error parsing JSON body", http.StatusBadRequest)
		reject = true
	}

	// Serialize and write to a log file
	d.writeToLogFile(reject, distro, variant, version, metrics)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("POST request processed\n"))
}

// validateVersion verifies the the version of Ubuntu is valid
// it's broader than the standard Ubuntu version for flavours
// that amy use slightly different numbering
func validateVersion(s string) bool {
	pattern := `^\d\d\.(0[1-9]|1[0-2])$`
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		slog.Error(fmt.Sprintf("Error compiling regex: %v", err))
		return false
	}

	return matched
}
