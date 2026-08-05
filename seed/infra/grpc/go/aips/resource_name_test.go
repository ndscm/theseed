package aips

import (
	"reflect"
	"testing"
)

func TestParseResourceName(t *testing.T) {
	t.Run("single pair", func(t *testing.T) {
		rn, err := ParseResourceName("people/123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("multiple pairs", func(t *testing.T) {
		rn, err := ParseResourceName("people/123/brainSteps/456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123", "brainSteps": "456"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("leading slash", func(t *testing.T) {
		rn, err := ParseResourceName("/people/123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("trailing slash", func(t *testing.T) {
		rn, err := ParseResourceName("people/123/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("surrounding whitespace", func(t *testing.T) {
		rn, err := ParseResourceName("  people/123  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("full name with authority", func(t *testing.T) {
		// AIP-122 full resource name: the authority is the service, not an id.
		rn, err := ParseResourceName("//library.googleapis.com/shelves/1/books/2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"shelves": "1", "books": "2"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("scheme and authority", func(t *testing.T) {
		rn, err := ParseResourceName("https://host/people/123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("escaped slash in id", func(t *testing.T) {
		rn, err := ParseResourceName("people/a%2Fb/steps/456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "a/b", "steps": "456"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("escaped colon in id", func(t *testing.T) {
		rn, err := ParseResourceName("people/a%3Ab")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "a:b"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("opaque scheme drops scheme", func(t *testing.T) {
		// A "scheme:opaque" name parses via Opaque; the scheme is dropped.
		rn, err := ParseResourceName("urn:people/123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := map[string]string{"people": "123"}
		if !reflect.DeepEqual(rn.Ids, expected) {
			t.Fatalf("ids = %v, expected %v", rn.Ids, expected)
		}
	})

	t.Run("empty name yields no ids", func(t *testing.T) {
		rn, err := ParseResourceName("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rn.Ids) != 0 {
			t.Fatalf("ids = %v, expected empty", rn.Ids)
		}
	})

	t.Run("slash only yields no ids", func(t *testing.T) {
		rn, err := ParseResourceName("/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rn.Ids) != 0 {
			t.Fatalf("ids = %v, expected empty", rn.Ids)
		}
	})

	t.Run("odd segment count errors", func(t *testing.T) {
		_, err := ParseResourceName("people/123/brainSteps")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("empty segment errors", func(t *testing.T) {
		_, err := ParseResourceName("people//brainSteps/456")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
