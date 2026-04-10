package tool

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMiddleware_chainOrder(t *testing.T) {
	reg := NewRegistry()
	var order []string
	reg.Use(func(next Handler) Handler {
		return func(ctx context.Context, args json.RawMessage) (string, error) {
			order = append(order, "outer")
			return next(ctx, args)
		}
	})
	reg.Use(func(next Handler) Handler {
		return func(ctx context.Context, args json.RawMessage) (string, error) {
			order = append(order, "inner")
			return next(ctx, args)
		}
	})
	_ = reg.Register(Definition{
		Name:       "x",
		Parameters: json.RawMessage(`{"type":"object"}`),
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		order = append(order, "handler")
		return "ok", nil
	})
	_, err := reg.Execute(context.Background(), "x", `{}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 || order[0] != "outer" || order[1] != "inner" || order[2] != "handler" {
		t.Fatalf("got %v", order)
	}
}

func TestPolicy_block(t *testing.T) {
	reg := NewRegistry()
	reg.UsePolicy(BlockToolNames("bad"))
	_ = reg.Register(Definition{
		Name:       "bad",
		Parameters: json.RawMessage(`{"type":"object"}`),
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "no", nil
	})
	_, err := reg.Execute(context.Background(), "bad", `{}`, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
