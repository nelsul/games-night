package games

import (
	"context"
	"testing"
)

func TestIsHost(t *testing.T) {
	ctx := WithHost(context.Background(), "ana")
	if !IsHost(ctx, "ana") || IsHost(ctx, "bia") {
		t.Fatal("only the host is the host")
	}
	if IsHost(context.Background(), "") {
		t.Fatal("no host in a plain context")
	}
}
