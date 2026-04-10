package mcp

import (
	"context"
	"strconv"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lethuan127/centrai-agent/internal/tool"
)

func TestRegisterRemoteTools_inMemory(t *testing.T) {
	ctx := context.Background()

	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "srv", Version: "v1"}, nil)
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "add",
		Description: "add two ints",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"type": "number"},
				"b": map[string]any{"type": "number"},
			},
			"required": []any{"a", "b"},
		},
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args map[string]any) (*mcpsdk.CallToolResult, any, error) {
		ai, _ := args["a"].(float64)
		bi, _ := args["b"].(float64)
		sum := int(ai + bi)
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: strconv.Itoa(sum)}},
		}, nil, nil
	})

	ct, st := mcpsdk.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "cli", Version: "v1"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	reg := tool.NewRegistry()
	if err := RegisterRemoteTools(ctx, session, reg, "mcp_"); err != nil {
		t.Fatal(err)
	}
	out, err := reg.Execute(ctx, "mcp_add", `{"a":2,"b":3}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	if out != "5" {
		t.Fatalf("unexpected result %q", out)
	}
}
