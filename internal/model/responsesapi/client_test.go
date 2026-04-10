package responsesapi

import (
	"testing"

	"github.com/lethuan127/centrai-agent/internal/model"
)

func TestClient_implementsModel(t *testing.T) {
	var _ model.Client = (*Client)(nil)
}
