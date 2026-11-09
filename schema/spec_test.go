package schema_test

// This file keeps the AI-facing canonical spec (gpdf-json-spec.{en,ja}.md)
// honest. It checks three things and regenerates the spec from its sources:
//
//  1. gpdf.schema.json matches the Go types that template.FromJSON decodes
//     into (no field missing from either side).
//  2. Every example in examples/ passes gpdf.schema.json AND renders through
//     template.FromJSON + Generate without error. Examples are what an LLM
//     copies first, so an example that does not pass is a lie in the spec.
//  3. The committed gpdf-json-spec.*.md files are byte-identical to what
//     spec/*.md + gpdf.schema.json + examples/ produce.
//
// Regenerate after editing any source:
//
//	go test ./schema -run TestSpec -update

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/template"
)

var update = flag.Bool("update", false, "rewrite gpdf-json-spec.*.md")

const schemaFile = "gpdf.schema.json"

func loadSchema(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var s map[string]any
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	return s
}

// ---------------------------------------------------------------------------
// 1. Schema ↔ Go types
// ---------------------------------------------------------------------------

// TestSchemaMatchesGoTypes walks template.Schema by reflection alongside the
// JSON Schema and fails when a JSON field exists on one side only.
func TestSchemaMatchesGoTypes(t *testing.T) {
	s := loadSchema(t)
	defs, _ := s["$defs"].(map[string]any)
	seen := map[reflect.Type]bool{}
	compareType(t, reflect.TypeOf(template.Schema{}), s, defs, "$", seen)
}

func compareType(t *testing.T, typ reflect.Type, node, defs map[string]any, path string, seen map[reflect.Type]bool) {
	t.Helper()
	node = deref(node, defs)
	if seen[typ] {
		return
	}
	seen[typ] = true

	props, _ := node["properties"].(map[string]any)
	goFields := map[string]reflect.Type{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		goFields[name] = f.Type
	}
	for name, ft := range goFields {
		p, ok := props[name].(map[string]any)
		if !ok {
			t.Errorf("%s.%s: decoded by %s but missing from %s", path, name, typ.Name(), schemaFile)
			continue
		}
		if st := structOf(ft); st != nil {
			child := p
			if items, ok := deref(p, defs)["items"].(map[string]any); ok {
				child = items
				// [][]T is not used for structs, so one level is enough.
			}
			compareType(t, st, child, defs, path+"."+name, seen)
		}
	}
	for name := range props {
		if _, ok := goFields[name]; !ok {
			t.Errorf("%s.%s: declared in %s but %s has no such field (gpdf would ignore it)", path, name, schemaFile, typ.Name())
		}
	}
}

func structOf(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t
	}
	return nil
}

func deref(node, defs map[string]any) map[string]any {
	for {
		ref, ok := node["$ref"].(string)
		if !ok {
			return node
		}
		next, ok := defs[strings.TrimPrefix(ref, "#/$defs/")].(map[string]any)
		if !ok {
			return node
		}
		node = next
	}
}

// ---------------------------------------------------------------------------
// 2. Examples
// ---------------------------------------------------------------------------

type example struct {
	name string
	path string
	data []byte
}

func loadExamples(t *testing.T) []example {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("examples", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 3 {
		t.Fatalf("examples/ has %d files; lost the directory?", len(paths))
	}
	sort.Strings(paths)
	out := make([]example, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, example{name: filepath.Base(p), path: p, data: data})
	}
	return out
}

// TestExamplesPassSchema validates every example against gpdf.schema.json.
func TestExamplesPassSchema(t *testing.T) {
	s := loadSchema(t)
	v := &validator{defs: s["$defs"].(map[string]any)}
	for _, ex := range loadExamples(t) {
		t.Run(ex.name, func(t *testing.T) {
			var doc any
			if err := json.Unmarshal(ex.data, &doc); err != nil {
				t.Fatalf("not JSON: %v", err)
			}
			for _, e := range v.validate(s, doc, "$") {
				t.Error(e)
			}
		})
	}
}

// hasCJK reports whether the example contains CJK text. Such examples are
// rendered the way the spec tells authors to: with a CJK TrueType font
// registered as the default font (tables have no per-cell fontFamily).
func hasCJK(data []byte) bool {
	for _, r := range string(data) {
		if r >= 0x2E80 {
			return true
		}
	}
	return false
}

// TestExamplesRender renders every example through template.FromJSON.
// CJK examples are rendered only when NotoSansJP-Regular.ttf is present at
// the workspace root (it is not committed to this repository).
func TestExamplesRender(t *testing.T) {
	for _, ex := range loadExamples(t) {
		t.Run(ex.name, func(t *testing.T) {
			var opts []template.Option
			if hasCJK(ex.data) {
				font, err := os.ReadFile(filepath.Join("..", "..", "NotoSansJP-Regular.ttf"))
				if err != nil {
					t.Skip("NotoSansJP-Regular.ttf not found; schema check still runs in TestExamplesPassSchema")
				}
				opts = append(opts,
					template.WithFont("NotoSansJP", font),
					template.WithDefaultFont("NotoSansJP", 10))
			}
			doc, err := template.FromJSON(ex.data, nil, opts...)
			if err != nil {
				t.Fatalf("FromJSON: %v", err)
			}
			pdf, err := doc.Generate()
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
				t.Fatal("output is not a PDF")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Canonical spec
// ---------------------------------------------------------------------------

// TestSpecUpToDate rebuilds gpdf-json-spec.{en,ja}.md and compares them
// with the committed files.
func TestSpecUpToDate(t *testing.T) {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatal(err)
	}
	examples := loadExamples(t)
	for _, lang := range []string{"en", "ja"} {
		prose, err := os.ReadFile(filepath.Join("spec", lang+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if err := checkProse(string(prose)); err != nil {
			t.Errorf("spec/%s.md: %v", lang, err)
		}
		got := buildSpec(lang, string(prose), schema, examples)
		out := fmt.Sprintf("gpdf-json-spec.%s.md", lang)
		if *update {
			if err := os.WriteFile(out, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(out)
		if err != nil || string(want) != got {
			t.Errorf("%s is stale; run: go test ./schema -run TestSpec -update", out)
		}
	}
}

// checkProse rejects hand-written JSON documents in the prose. Full
// documents belong in examples/ where they are validated; prose may only
// show fragments (which never start with "page").
func checkProse(prose string) error {
	if strings.Contains(prose, "```json\n{\n  \"page\"") {
		return fmt.Errorf("contains a full JSON document; move it to examples/")
	}
	return nil
}

var specText = map[string]map[string]string{
	"en": {
		"generated": "<!-- Generated by `go test ./schema -run TestSpec -update`. Do not edit; edit schema/spec/en.md, schema/gpdf.schema.json or schema/examples/ instead. -->",
		"appA":      "## Appendix A. JSON Schema",
		"appALede":  "This is the only source of truth for field names, types, enums and patterns. The prose above does not repeat them.",
		"appB":      "## Appendix B. Examples",
		"appBLede":  "Every example below passes Appendix A and renders through `template.FromJSON` in gpdf's test suite. Start from the closest one.",
	},
	"ja": {
		"generated": "<!-- `go test ./schema -run TestSpec -update` で生成。手で編集しない。schema/spec/ja.md・schema/gpdf.schema.json・schema/examples/ を直すこと。 -->",
		"appA":      "## 付録A. JSON Schema",
		"appALede":  "欄の名前・型・列挙値・書式の正はここだけ。本文には写していない。",
		"appB":      "## 付録B. 実例",
		"appBLede":  "以下の実例はすべて、gpdf のテストで付録A の検証と `template.FromJSON` による生成を通している。いちばん近いものから書き始めること。",
	},
}

func buildSpec(lang, prose string, schema []byte, examples []example) string {
	txt := specText[lang]
	var b strings.Builder
	b.WriteString(txt["generated"])
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(prose))
	b.WriteString("\n\n---\n\n")
	b.WriteString(txt["appA"] + "\n\n" + txt["appALede"] + "\n\n")
	b.WriteString("#### `gpdf.schema.json`\n\n```json\n")
	b.WriteString(strings.TrimRight(string(schema), "\n"))
	b.WriteString("\n```\n\n")
	b.WriteString(txt["appB"] + "\n\n" + txt["appBLede"] + "\n")
	for _, ex := range examples {
		b.WriteString("\n#### `" + ex.name + "`\n\n```json\n")
		b.WriteString(strings.TrimRight(string(ex.data), "\n"))
		b.WriteString("\n```\n")
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Minimal JSON Schema validator
// ---------------------------------------------------------------------------

// validator implements the subset of JSON Schema 2020-12 that
// gpdf.schema.json uses. An unsupported keyword fails the test so the
// subset cannot silently fall behind the schema.
type validator struct {
	defs map[string]any
}

var annotationKeys = map[string]bool{
	"$schema": true, "$id": true, "$defs": true, "title": true,
	"description": true, "default": true,
}

func (v *validator) validate(node map[string]any, val any, path string) []string {
	var errs []string
	// "type" first: when it fails, the other keywords would only add noise.
	if typ, ok := node["type"].(string); ok && !hasType(val, typ) {
		return []string{fmt.Sprintf("%s: want %s, got %T", path, typ, val)}
	}
	for key, raw := range node {
		if key == "type" || annotationKeys[key] {
			continue
		}
		check, ok := keywordChecks[key]
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: validator does not support keyword %q; extend schema/spec_test.go", path, key))
			continue
		}
		errs = append(errs, check(v, node, raw, val, path)...)
	}
	return errs
}

type keywordCheck func(v *validator, node map[string]any, raw, val any, path string) []string

var keywordChecks map[string]keywordCheck

func init() {
	keywordChecks = map[string]keywordCheck{
		"$ref":                 checkRef,
		"enum":                 checkEnum,
		"pattern":              checkPattern,
		"anyOf":                checkAnyOf,
		"minimum":              checkBound("minimum", func(n, lim float64) bool { return n >= lim }),
		"maximum":              checkBound("maximum", func(n, lim float64) bool { return n <= lim }),
		"exclusiveMinimum":     checkBound("exclusiveMinimum", func(n, lim float64) bool { return n > lim }),
		"required":             checkRequired,
		"properties":           checkProperties,
		"additionalProperties": checkAdditional,
		"items":                checkItems,
		"minItems":             checkCount("minItems", func(n, lim int) bool { return n >= lim }),
		"maxItems":             checkCount("maxItems", func(n, lim int) bool { return n <= lim }),
	}
}

func fail(path, format string, args ...any) []string {
	return []string{path + ": " + fmt.Sprintf(format, args...)}
}

func checkRef(v *validator, _ map[string]any, raw, val any, path string) []string {
	def, ok := v.defs[strings.TrimPrefix(raw.(string), "#/$defs/")].(map[string]any)
	if !ok {
		return fail(path, "unresolved $ref %q", raw)
	}
	return v.validate(def, val, path)
}

func checkEnum(_ *validator, _ map[string]any, raw, val any, path string) []string {
	for _, e := range raw.([]any) {
		if reflect.DeepEqual(e, val) {
			return nil
		}
	}
	return fail(path, "%v is not one of %v", val, raw)
}

func checkPattern(_ *validator, _ map[string]any, raw, val any, path string) []string {
	if s, ok := val.(string); ok && !regexp.MustCompile(raw.(string)).MatchString(s) {
		return fail(path, "%q does not match %s", s, raw)
	}
	return nil
}

func checkAnyOf(v *validator, _ map[string]any, raw, val any, path string) []string {
	for _, sub := range raw.([]any) {
		if len(v.validate(sub.(map[string]any), val, path)) == 0 {
			return nil
		}
	}
	return fail(path, "%v matches none of anyOf", val)
}

func checkBound(key string, ok func(n, lim float64) bool) keywordCheck {
	return func(_ *validator, _ map[string]any, raw, val any, path string) []string {
		if n, isNum := val.(float64); isNum && !ok(n, raw.(float64)) {
			return fail(path, "%v violates %s %v", n, key, raw)
		}
		return nil
	}
}

func checkCount(key string, ok func(n, lim int) bool) keywordCheck {
	return func(_ *validator, _ map[string]any, raw, val any, path string) []string {
		if arr, isArr := val.([]any); isArr && !ok(len(arr), int(raw.(float64))) {
			return fail(path, "%d items violates %s %v", len(arr), key, raw)
		}
		return nil
	}
}

func checkRequired(_ *validator, _ map[string]any, raw, val any, path string) []string {
	obj, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	var errs []string
	for _, r := range raw.([]any) {
		if _, ok := obj[r.(string)]; !ok {
			errs = append(errs, fail(path, "missing required field %q", r)...)
		}
	}
	return errs
}

func checkProperties(v *validator, _ map[string]any, raw, val any, path string) []string {
	obj, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	var errs []string
	props := raw.(map[string]any)
	for k, fv := range obj {
		if p, ok := props[k].(map[string]any); ok {
			errs = append(errs, v.validate(p, fv, path+"."+k)...)
		}
	}
	return errs
}

func checkAdditional(_ *validator, node map[string]any, raw, val any, path string) []string {
	obj, ok := val.(map[string]any)
	if !ok || raw != false {
		return nil
	}
	props, _ := node["properties"].(map[string]any)
	var errs []string
	for k := range obj {
		if _, ok := props[k]; !ok {
			errs = append(errs, fail(path, "unknown field %q", k)...)
		}
	}
	return errs
}

func checkItems(v *validator, _ map[string]any, raw, val any, path string) []string {
	arr, ok := val.([]any)
	if !ok {
		return nil
	}
	var errs []string
	for i, item := range arr {
		errs = append(errs, v.validate(raw.(map[string]any), item, fmt.Sprintf("%s[%d]", path, i))...)
	}
	return errs
}

func hasType(val any, typ string) bool {
	switch typ {
	case "object":
		_, ok := val.(map[string]any)
		return ok
	case "array":
		_, ok := val.([]any)
		return ok
	case "string":
		_, ok := val.(string)
		return ok
	case "number":
		_, ok := val.(float64)
		return ok
	case "integer":
		n, ok := val.(float64)
		return ok && n == math.Trunc(n)
	case "boolean":
		_, ok := val.(bool)
		return ok
	}
	return false
}

// TestValidatorRejects guards the validator itself: each document must
// produce at least one error.
func TestValidatorRejects(t *testing.T) {
	s := loadSchema(t)
	v := &validator{defs: s["$defs"].(map[string]any)}
	bad := map[string]string{
		"unknown top-level field": `{"page":{"size":"A4"},"defaultFont":{"family":"X"}}`,
		"missing page":            `{"body":[]}`,
		"bad color":               `{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":12,"text":"a","style":{"color":"#ggg"}}]}}]}`,
		"bad dimension":           `{"page":{"size":"A4","margins":"1px"}}`,
		"bad page size":           `{"page":{"size":"A99"}}`,
		"span out of range":       `{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":13,"text":"a"}]}}]}`,
		"unknown element type":    `{"page":{"size":"A4"},"body":[{"row":{"cols":[{"span":12,"elements":[{"type":"paragraph"}]}]}}]}`,
	}
	for name, js := range bad {
		var doc any
		if err := json.Unmarshal([]byte(js), &doc); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if errs := v.validate(s, doc, "$"); len(errs) == 0 {
			t.Errorf("%s: validator accepted %s", name, js)
		}
	}
}
