package gengroovy_test

import (
	"testing"

	"github.com/ndscm/theseed/seed/infra/groovy"
	"github.com/ndscm/theseed/seed/infra/groovy/generator/gengroovy"
)

func TestGenerateFromGroovy(t *testing.T) {
	t.Run("package_imports_and_class_members", func(t *testing.T) {
		source := `package app.demo

import java.util.List
import static java.lang.Math.max as maximum

class A extends B implements C, D {
  def x = [1, 2, 3]
  private int y = 2
  static {
    boot()
  }
  A(String name) {
    return
  }
  def run(int n = 1) {
    call(n)
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

	t.Run("script_statements", func(t *testing.T) {
		source := `total = 1 + 2
println(total)
throw error
break done
continue next`
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

	t.Run("top_level_method_with_list_and_map_literals", func(t *testing.T) {
		source := `def build() {
  values([1, 2], [name: 'Ada', active: true])
  return result
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

	t.Run("interface_extends_and_static_star_import", func(t *testing.T) {
		source := `import static java.util.Collections.*

interface Worker extends Runnable, Closeable {
  String name = 'seed'
  def run() {
    return
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

	t.Run("custom_indent", func(t *testing.T) {
		source := `class A {
  def x = 1
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
}
