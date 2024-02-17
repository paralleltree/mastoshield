package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/paralleltree/mastoshield/config"
	"github.com/paralleltree/mastoshield/rule"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
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
		log.Fatalf("%v", err)
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
	upstreamUrl, err := url.Parse(conf.UpstreamEndpoint)
	if err != nil {
		return fmt.Errorf("parse upstream url: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstreamUrl)
	addr := fmt.Sprintf(":%d", conf.ListenPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/", Handler(proxy, conf.DenyResponseCode, rulesets))
	server := &http.Server{Addr: addr, Handler: mux}

	fmt.Println("starting")

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
	return nil
}

func Handler(upstream http.Handler, denyResponseCode int, rulesets []rule.RuleSet) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var readBody []byte
		bodyFetcher := func() ([]byte, error) {
			if readBody != nil {
				return readBody, nil
			}
			defer r.Body.Close()
			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			readBody = rawBody
			r.Body = io.NopCloser(bytes.NewBuffer(readBody))
			return readBody, nil
		}

		testRequest := func(r *http.Request, ruleset rule.RuleSet) (bool, error) {
			for _, matcher := range ruleset.Matchers {
				matched, err := matcher.Test(r, bodyFetcher)
				if err != nil {
					return false, fmt.Errorf("test request: %w", err)
				}
				if !matched {
					return false, nil
				}
			}
			return true, nil
		}

		for _, ruleset := range rulesets {
			matched, err := testRequest(r, ruleset)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte{})
				return
			}
			if matched {
				switch ruleset.Action {
				case rule.ACTION_ALLOW:
					// call callback func to logging
					// TODO: logging
					upstream.ServeHTTP(w, r)
				case rule.ACTION_DENY:
					w.WriteHeader(denyResponseCode)
					w.Write([]byte{})
					b, _ := bodyFetcher()
					fmt.Printf("DENIED (%s)\n", string(b))
				default:
					// TODO: logging
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte{})
				}
				return
			}
		}

		// default action(allow)
		upstream.ServeHTTP(w, r)
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
