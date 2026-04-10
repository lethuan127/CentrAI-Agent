package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lethuan127/centrai-agent/internal/agent"
	"github.com/lethuan127/centrai-agent/internal/agentdef"
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
		var err error
		def, err = agentdef.LoadFile(cfg.AgentFile)
		if err != nil {
			slog.Error("load agent yaml", "file", cfg.AgentFile, "err", err)
			os.Exit(2)
		}
		if n := strings.TrimSpace(def.Name); n != "" {
			slog.Info("agent definition", "name", n, "file", cfg.AgentFile)
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
	if def != nil && def.MaxSteps != nil && *def.MaxSteps > 0 {
		maxSteps = *def.MaxSteps
	}

	runner := agent.NewRunner(store, m, reg, agent.Options{
		MaxSteps: maxSteps,
		Logger:   agent.NewSlogLogger(slog.Default()),
	})

	system := buildSystemPrompt(def, wantDemo)

	ctx := context.Background()

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
		slog.Error("pass -message TEXT or use -repl")
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
