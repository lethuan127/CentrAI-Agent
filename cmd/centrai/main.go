// Command centrai is the reference CLI for this module. It reads configuration from
// flags and environment only (see config.go), wires the agent runner, optional HTTP
// ingress ([api/openapi.yaml](../../api/openapi.yaml)), and REPL or one-shot modes.
// Documentation: [docs/plan.md](../../docs/plan.md), [docs/9. code-structure.md](../../docs/9.%20code-structure.md).
package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/agentdef"
	"github.com/lethuan127/centrai-agent/internal/httpserver"
	"github.com/lethuan127/centrai-agent/internal/model/openai"
	"github.com/lethuan127/centrai-agent/internal/store/memory"
	"github.com/lethuan127/centrai-agent/internal/tool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := loadConfig()
	if cfg.APIKey == "" && os.Getenv("CENTRAI_SKIP_API") == "" {
		slog.Error("OPENAI_API_KEY is not set (export CENTRAI_SKIP_API=1 to run without a live model)")
		os.Exit(2)
	}

	var def *agentdef.Definition
	if cfg.AgentFile != "" {
		agentPath := resolveAgentFile(cfg.AgentFile)
		if agentPath != cfg.AgentFile {
			slog.Info("resolved -agent", "arg", cfg.AgentFile, "file", agentPath)
		}
		var err error
		def, err = agentdef.LoadFile(agentPath)
		if err != nil {
			slog.Error("load agent definition", "file", agentPath, "err", err)
			os.Exit(2)
		}
		if n := strings.TrimSpace(def.Name); n != "" {
			slog.Info("agent definition", "name", n, "file", agentPath)
		}
		if p := strings.TrimSpace(def.Provider); p != "" && !strings.EqualFold(p, "openai") {
			slog.Warn("agent provider field set but CLI uses OpenAI-compatible client only for now", "provider", p)
		}
		if len(def.McpServers) > 0 {
			slog.Info("agent lists MCP servers (wire with github.com/modelcontextprotocol/go-sdk and internal/mcp.RegisterRemoteTools)", "mcpServers", def.McpServers)
		}
		if len(def.Skills) > 0 {
			slog.Info("agent lists skills (see internal/skill.Loader)", "skills", def.Skills)
		}
	}

	wantDemo := cfg.DemoTools || (def != nil && def.WantsDemoTools())

	store := memory.New()
	reg := tool.NewRegistry()
	if wantDemo {
		if err := registerDemoTools(reg); err != nil {
			slog.Error("demo tools", "err", err)
			os.Exit(2)
		}
	}

	modelName := cfg.Model
	if def != nil && strings.TrimSpace(def.Model) != "" {
		modelName = strings.TrimSpace(def.Model)
	}

	m := openai.New(openai.Config{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   modelName,
	})

	maxSteps := cfg.MaxSteps
	if def != nil {
		maxSteps = def.TurnLimit(cfg.MaxSteps)
	}

	runner := agent.NewRunner(store, m, reg, agent.Options{
		MaxSteps: maxSteps,
		Logger:   agent.NewSlogLogger(slog.Default()),
	})

	system := buildSystemPrompt(def, wantDemo)

	ctx := context.Background()

	if cfg.HTTPAddr != "" {
		h := httpserver.Handler(httpserver.Options{Runner: runner, DefaultSystem: system})
		srv := &http.Server{Addr: cfg.HTTPAddr, Handler: h, ReadHeaderTimeout: 10 * time.Second}
		sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go func() {
			slog.Info("http listening", "addr", cfg.HTTPAddr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("http server", "err", err)
				os.Exit(1)
			}
		}()
		<-sigCtx.Done()
		sdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(sdCtx)
		return
	}

	if cfg.Repl {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			out, err := runner.Run(ctx, &agent.RunInput{
				SessionID:          cfg.Session,
				UserMessage:        line,
				System:             system,
				Model:              modelName,
				MaxSteps:           maxSteps,
				OmitSessionContext: cfg.OmitSessionContext,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			fmt.Println(out.Assistant)
		}
		if err := sc.Err(); err != nil {
			slog.Error("stdin", "err", err)
			os.Exit(1)
		}
		return
	}

	if cfg.Message == "" {
		slog.Error("pass -message TEXT, use -repl, or -http ADDR")
		os.Exit(2)
	}

	out, err := runner.Run(ctx, &agent.RunInput{
		SessionID:          cfg.Session,
		UserMessage:        cfg.Message,
		System:             system,
		Model:              modelName,
		MaxSteps:           maxSteps,
		OmitSessionContext: cfg.OmitSessionContext,
	})
	if err != nil {
		slog.Error("run failed", "err", err)
		os.Exit(1)
	}
	fmt.Println(out.Assistant)
}
