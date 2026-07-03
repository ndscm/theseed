package servicetier

import (
	"testing"
)

func TestStaticParserParse(t *testing.T) {
	parser := NewStaticParser(
		[]string{"ndscm.com", "ndscm.biz"},
	)

	t.Run("go.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("stuff.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("stuff.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "stuff",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.prod.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "prod",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.staging.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.staging.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "staging",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.future.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.future.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "future",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.prod.svc.sfo.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.prod.svc.sfo.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "prod",
			Role:       "svc",
			Cluster:    "sfo",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.dev-christina.svc.sfo.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.dev-christina.svc.sfo.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "dev",
			Owner:      "christina",
			Role:       "svc",
			Cluster:    "sfo",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.dev-christina.svc.sfocorp.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.dev-christina.svc.sfocorp.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "dev",
			Owner:      "christina",
			Role:       "svc",
			Cluster:    "sfocorp",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("go.dev-christina.svc.local.ndscm.biz", func(t *testing.T) {
		got, err := parser.Parse("go.dev-christina.svc.local.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "dev",
			Owner:      "christina",
			Role:       "svc",
			Cluster:    "local",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("three parts sets cluster not role", func(t *testing.T) {
		got, err := parser.Parse("go.dev-christina.sfo.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "dev",
			Owner:      "christina",
			Cluster:    "sfo",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("ndscm.com root domain", func(t *testing.T) {
		got, err := parser.Parse("go.ndscm.com")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.com",
			Service:    "go",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("port is parsed", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.biz:8080")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "prod",
			Port:       "8080",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("bare root domain has no service", func(t *testing.T) {
		got, err := parser.Parse("ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("single label host returns early", func(t *testing.T) {
		got, err := parser.Parse("localhost")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "localhost",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("single label host with port", func(t *testing.T) {
		got, err := parser.Parse("localhost:8080")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "localhost",
			Port:       "8080",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("ipv4 host returns early", func(t *testing.T) {
		got, err := parser.Parse("127.0.0.1")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "127.0.0.1",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("ipv4 host with port", func(t *testing.T) {
		got, err := parser.Parse("127.0.0.1:8080")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "127.0.0.1",
			Port:       "8080",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("ipv6 host returns early", func(t *testing.T) {
		got, err := parser.Parse("::1")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "::1",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("ipv6 host with port", func(t *testing.T) {
		got, err := parser.Parse("[::1]:8080")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "::1",
			Port:       "8080",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("unknown root domain errors", func(t *testing.T) {
		got, err := parser.Parse("go.example.com")
		if err == nil {
			t.Fatalf("Parse: expected error, got %+v", got)
		}
	})

	t.Run("too many host parts errors", func(t *testing.T) {
		got, err := parser.Parse("a.b.c.d.e.ndscm.biz")
		if err == nil {
			t.Fatalf("Parse: expected error, got %+v", got)
		}
	})

	t.Run("root domain host parses", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "prod",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("tenant domain host errors without tenant verifier", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.ndteam.net")
		if err == nil {
			t.Fatalf("Parse: expected error, got %+v", got)
		}
	})
}

func TestStaticMultiTenantParserParse(t *testing.T) {
	parser := NewStaticMultiTenantParser(
		[]string{"ndscm.com", "ndscm.biz"},
		[]string{"ndteam.net"},
	)

	t.Run("root domain still parses", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.biz")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndscm.biz",
			Service:    "go",
			Tier:       "prod",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("tenant domain extracts tenant", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.ndteam.net")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndteam.net",
			Tenant:     "ndscm",
			Service:    "go",
			Tier:       "prod",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("tenant with only service", func(t *testing.T) {
		got, err := parser.Parse("go.ndscm.ndteam.net")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndteam.net",
			Tenant:     "ndscm",
			Service:    "go",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("tenant with role and cluster", func(t *testing.T) {
		got, err := parser.Parse("go.dev-christina.svc.sfo.ndscm.ndteam.net")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndteam.net",
			Tenant:     "ndscm",
			Service:    "go",
			Tier:       "dev",
			Owner:      "christina",
			Role:       "svc",
			Cluster:    "sfo",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("tenant domain with port", func(t *testing.T) {
		got, err := parser.Parse("go.prod.ndscm.ndteam.net:8080")
		if err != nil {
			t.Fatalf("Parse: unexpected error: %v", err)
		}
		want := ServiceTier{
			RootDomain: "ndteam.net",
			Tenant:     "ndscm",
			Service:    "go",
			Tier:       "prod",
			Port:       "8080",
		}
		if *got != want {
			t.Errorf("Parse mismatch:\n got  %+v\n want %+v", *got, want)
		}
	})

	t.Run("unknown domain errors", func(t *testing.T) {
		got, err := parser.Parse("go.example.com")
		if err == nil {
			t.Fatalf("Parse: expected error, got %+v", got)
		}
	})
}
