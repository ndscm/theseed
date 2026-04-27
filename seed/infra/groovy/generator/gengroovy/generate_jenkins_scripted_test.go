package gengroovy_test

import (
	"testing"

	"github.com/ndscm/theseed/seed/infra/groovy"
	"github.com/ndscm/theseed/seed/infra/groovy/generator/gengroovy"
)

func TestGenerateFromJenkinsScriptedPipeline(t *testing.T) {
	t.Run("node_with_stage_steps", func(t *testing.T) {
		source := `node('linux') {
  stage('Checkout') {
    checkout(scm)
  }
  stage('Test') {
    sh('make test')
  }
}`
		expected := source
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})

	t.Run("properties_parallel_and_archive", func(t *testing.T) {
		source := `properties([disableConcurrentBuilds()])
parallel(
  linux: {
    sh('make test-linux')
  },
  darwin: {
    sh('make test-darwin')
  }
)
archiveArtifacts('dist/**')`
		expected := source
		module, err := groovy.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gengroovy.GenerateIndent(module, "  ")
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Fatalf("GenerateIndent() =\n%q\nexpected\n%q", got, expected)
		}
	})
}
