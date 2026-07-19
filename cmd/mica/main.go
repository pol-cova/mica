package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mica-dev/mica/cmd/mica/web"
	"github.com/mica-dev/mica/internal/api"
	"github.com/mica-dev/mica/internal/catalog"
	"github.com/mica-dev/mica/internal/incidents"
	"github.com/mica-dev/mica/internal/mcpserver"
	"github.com/mica-dev/mica/internal/prometheus"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	dataDir := os.Getenv("MICA_DATA_DIR")
	if dataDir == "" {
		dataDir = ".mica"
	}

	switch os.Args[1] {
	case "serve":
		serve(dataDir, os.Args[2:])
	case "demo":
		demo(dataDir, os.Args[2:])
	case "mcp":
		mcp(dataDir)
	default:
		usage()
		os.Exit(2)
	}
}

func mcp(dataDir string) {
	store, err := incidents.Open(filepath.Join(dataDir, "mica.db"))
	if err != nil {
		slog.Error("open incident store", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := loadCatalog(store); err != nil {
		slog.Error("load service catalog", "error", err)
		os.Exit(1)
	}
	if url := os.Getenv("MICA_PROMETHEUS_URL"); url != "" {
		store.SetMetricsSource(prometheus.Source{BaseURL: url, BearerToken: os.Getenv("MICA_PROMETHEUS_BEARER_TOKEN"), BasicUser: os.Getenv("MICA_PROMETHEUS_BASIC_USER"), BasicPass: os.Getenv("MICA_PROMETHEUS_BASIC_PASSWORD")})
	}
	if err := mcpserver.New(store).Serve(os.Stdin, os.Stdout); err != nil {
		slog.Error("MCP server", "error", err)
		os.Exit(1)
	}
}

func serve(dataDir string, args []string) {
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := flags.String("addr", "127.0.0.1:8787", "HTTP listen address")
	prometheusURL := flags.String("prometheus-url", "", "read-only Prometheus HTTP API base URL")
	prometheusToken := flags.String("prometheus-bearer-token", "", "Prometheus bearer token (prefer MICA_PROMETHEUS_BEARER_TOKEN)")
	prometheusUser := flags.String("prometheus-basic-user", "", "Prometheus basic-auth user (prefer MICA_PROMETHEUS_BASIC_USER)")
	prometheusPass := flags.String("prometheus-basic-password", "", "Prometheus basic-auth password (prefer MICA_PROMETHEUS_BASIC_PASSWORD)")
	flags.Parse(args)

	store, err := incidents.Open(filepath.Join(dataDir, "mica.db"))
	if err != nil {
		slog.Error("open incident store", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := loadCatalog(store); err != nil {
		slog.Error("load service catalog", "error", err)
		os.Exit(1)
	}
	if *prometheusURL != "" {
		token := *prometheusToken
		if token == "" {
			token = os.Getenv("MICA_PROMETHEUS_BEARER_TOKEN")
		}
		user, pass := *prometheusUser, *prometheusPass
		if user == "" {
			user = os.Getenv("MICA_PROMETHEUS_BASIC_USER")
		}
		if pass == "" {
			pass = os.Getenv("MICA_PROMETHEUS_BASIC_PASSWORD")
		}
		if token != "" && (user != "" || pass != "") {
			slog.Error("only one Prometheus authentication method may be configured")
			os.Exit(2)
		}
		source := prometheus.Source{BaseURL: *prometheusURL, BearerToken: token, BasicUser: user, BasicPass: pass}
		if window := os.Getenv("MICA_PROMETHEUS_RATE_WINDOW"); window != "" {
			source.Signals = prometheus.CheckoutSignals(window)
		}
		store.SetMetricsSource(source)
		slog.Info("using Prometheus metrics source", "url", *prometheusURL)
	}

	server := &http.Server{Addr: *addr, Handler: api.NewHandlerWithWeb(store, web.Assets())}
	go func() {
		slog.Info("Mica workspace ready", "url", "http://"+*addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("serve", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func loadCatalog(store *incidents.Store) error {
	path := os.Getenv("MICA_SERVICE_CATALOG")
	if path == "" {
		return nil
	}
	services, err := catalog.Load(path)
	if err != nil {
		return err
	}
	return store.ReplaceServices(services)
}

func demo(dataDir string, args []string) {
	if len(args) < 1 || len(args) > 2 || (args[0] != "trigger" && args[0] != "status" && args[0] != "reset") || (len(args) == 2 && (args[0] != "trigger" || args[1] != "n-plus-one")) {
		fmt.Fprintln(os.Stderr, "usage: mica demo trigger [n-plus-one] | mica demo status | mica demo reset")
		os.Exit(2)
	}
	if controlURL := os.Getenv("MICA_DEMO_CONTROL_URL"); controlURL != "" {
		if args[0] == "status" {
			enabled, err := demoRegressionStatus(controlURL)
			if err != nil {
				slog.Error("read checkout demo status", "error", err)
				os.Exit(1)
			}
			fmt.Printf("regression: %t\n", enabled)
			return
		}
		enabled := args[0] == "trigger"
		if err := setDemoRegression(controlURL, enabled); err != nil {
			slog.Error("set checkout demo regression", "error", err)
			os.Exit(1)
		}
	}
	store, err := incidents.Open(filepath.Join(dataDir, "mica.db"))
	if err != nil {
		slog.Error("open incident store", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if args[0] == "status" {
		fmt.Printf("regression: %t\n", store.DemoRegression())
		return
	}
	if args[0] == "trigger" {
		err = store.SetDemoRegression(true)
	} else {
		err = store.SetDemoRegression(false)
	}
	if err != nil {
		slog.Error("update demo", "error", err)
		os.Exit(1)
	}
	fmt.Printf("demo %s complete\n", args[0])
}

func demoRegressionStatus(baseURL string) (bool, error) {
	response, err := (&http.Client{Timeout: 5 * time.Second}).Get(baseURL + "/demo/status")
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return false, fmt.Errorf("demo control returned %s", response.Status)
	}
	var status struct {
		Regression bool `json:"regression"`
	}
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return false, err
	}
	return status.Regression, nil
}

func setDemoRegression(baseURL string, enabled bool) error {
	payload, err := json.Marshal(map[string]bool{"enabled": enabled})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, baseURL+"/demo/regression", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("demo control returned %s", response.Status)
	}
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mica <serve|demo|mcp>")
}
