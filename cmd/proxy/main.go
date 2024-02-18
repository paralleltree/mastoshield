package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hnakamur/ltsvlog/v3"
	"github.com/paralleltree/mastoshield/config"
	"github.com/paralleltree/mastoshield/rule"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	app := cli.App{
		Name: "mastoshield",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "rule-file",
				Required: true,
				Usage:    "Specify the yaml file including rules",
			},
		},
		Action: func(ctx *cli.Context) error {
			ruleFilePath := ctx.String("rule-file")
			return run(ctx.Context, ruleFilePath)
		},
	}
	if err := app.RunContext(ctx, os.Args); err != nil {
		ltsvlog.Logger.Err(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, ruleFilePath string) error {
	conf, err := config.LoadProxyConfig()
	if err != nil {
		return fmt.Errorf("load proxy config: %w", err)
	}
	rulesets, err := loadAccessControlConfig(ruleFilePath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := start(ctx, conf, rulesets); err != nil {
		return fmt.Errorf("running server: %w", err)
	}
	return nil
}

func start(ctx context.Context, conf *config.ProxyConfig, rulesets []rule.RuleSet) error {
	onAllowed := func(xid string, r *http.Request) {
		reportRequest(xid, r, "allowed")
	}
	onDenied := func(xid string, r *http.Request) {
		reportRequest(xid, r, "denied")
	}
	onError := func(xid string, err error) {
		ltsvlog.Logger.Err(fmt.Errorf("xid: %s: %v", xid, err))
	}
	upstreamUrl, err := url.Parse(conf.UpstreamEndpoint)
	if err != nil {
		return fmt.Errorf("parse upstream url: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstreamUrl)
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler(proxy, conf.DenyResponseCode, rulesets, nil, onAllowed, onDenied, onError))
	addr := fmt.Sprintf(":%d", conf.ListenPort)
	server := &http.Server{Addr: addr, Handler: mux}

	ltsvlog.Logger.Info().String("event", "start").Int("port", conf.ListenPort).String("upstream", conf.UpstreamEndpoint).Log()

	go func(ctx context.Context) {
		<-ctx.Done()
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.ExitTimeoutSeconds))
		defer cancel()
		if err := server.Shutdown(timeout); err != nil {
			log.Fatalf("shutdown server: %v", err)
		}
	}(ctx)

	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
	}
	ltsvlog.Logger.Info().String("event", "shutdown").Log()
	return nil
}

func reportRequest(xid string, r *http.Request, action string) {
	ltsvlog.Logger.Info().
		String("event", "requestHandled").
		String("xid", xid).
		String("action", action).
		String("method", r.Method).
		String("path", r.URL.Path).
		String("url", r.URL.String()).
		String("remote", resolveClientIP(r)).
		String("useragent", r.UserAgent()).
		Log()
}

func resolveClientIP(r *http.Request) string {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		addrs := strings.Split(prior, ",")
		if len(addrs) > 0 {
			return addrs[0]
		}
	}
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "-"
	}
	return clientIP
}

func Handler(
	upstream http.Handler, denyResponseCode int, rulesets []rule.RuleSet,
	onProcessing func(string, *http.Request), onAllowed func(string, *http.Request), onDenied func(string, *http.Request),
	onError func(string, error),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := xid.New().String()

		if onProcessing != nil {
			onProcessing(reqID, r)
		}

		testRequest := func(r *rule.ProxyRequest, ruleset rule.RuleSet) (bool, error) {
			for _, matcher := range ruleset.Matchers {
				matched, err := matcher.Test(r)
				if err != nil {
					return false, fmt.Errorf("test request: %w", err)
				}
				if !matched {
					return false, nil
				}
			}
			return true, nil
		}

		allowAction := func(w http.ResponseWriter, r *http.Request) {
			if onAllowed != nil {
				defer onAllowed(reqID, r)
			}
			upstream.ServeHTTP(w, r)
		}
		denyAction := func(w http.ResponseWriter, r *http.Request) {
			if onDenied != nil {
				defer onDenied(reqID, r)
			}
			w.WriteHeader(denyResponseCode)
			w.Write([]byte{})
		}
		errAction := func(w http.ResponseWriter, r *http.Request, err error) {
			if onError != nil {
				defer onError(reqID, err)
			}
			allowAction(w, r)
		}

		for _, ruleset := range rulesets {
			matched, err := testRequest(rule.NewProxyRequest(r), ruleset)
			if err != nil {
				errAction(w, r, err)
				return
			}
			if matched {
				switch ruleset.Action {
				case rule.ACTION_ALLOW:
					allowAction(w, r)
				case rule.ACTION_DENY:
					denyAction(w, r)
				default:
					errAction(w, r, err)
				}
				return
			}
		}

		// default action(allow)
		allowAction(w, r)
	}
}

func loadAccessControlConfig(path string) ([]rule.RuleSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	conf, err := config.LoadAccessControlConfig(f)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return conf, nil
}
