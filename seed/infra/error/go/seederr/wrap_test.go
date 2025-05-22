package seederr_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func TestWrapSimpleError(t *testing.T) {
	err := seederr.Wrap(fmt.Errorf("simple error"))
	if !strings.HasPrefix(fmt.Sprintf("%v", err), "simple error") {
		t.Errorf("Expected error to be wrapped, got: %v", err)
	}
}
