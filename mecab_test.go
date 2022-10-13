package mecab

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

var mecabrcPath string

func init() {
	// get mecabrc path from mecab-config from the environment value.
	if path := os.Getenv("MECABRC_PATH"); path != "" {
		mecabrcPath = path
	}
}

func rcfile(config map[string]string) map[string]string {
	if mecabrcPath != "" {
		config["rcfile"] = mecabrcPath
	}
	return config
}

func TestNewMeCab(t *testing.T) {
	mecab, err := New(rcfile(map[string]string{
		"output-format-type": "wakati",
		"all-morphs":         "",
	}))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer mecab.Destroy()
}

func TestNewMeCab_error(t *testing.T) {
	_, err := New(rcfile(map[string]string{
		"output-format-type": "unknown format",
	}))
	if err == nil {
		t.Errorf("expected error, but not")
		return
	}
	if !strings.Contains(err.Error(), "unknown format type [unknown format]") {
		t.Errorf("want %q error, got %q", "unknown format type [unknown format]", err.Error())
	}
}

func TestParse(t *testing.T) {
	mecab, err := New(rcfile(map[string]string{
		"output-format-type": "wakati",
	}))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer mecab.Destroy()

	result, err := mecab.Parse("こんにちは世界")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	expected := "こんにちは 世界 \n"
	if result != expected {
		t.Errorf("want `%s`, but `%s`", expected, result)
	}
}

func BenchmarkParse(b *testing.B) {
	mecab, _ := New(rcfile(map[string]string{
		"output-format-type": "wakati",
	}))
	defer mecab.Destroy()

	for i := 0; i < b.N; i++ {
		mecab.Parse("こんにちは世界")
	}
}

func TestParseLattice(t *testing.T) {
	mecab, err := New(rcfile(map[string]string{
		"output-format-type": "wakati",
	}))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer mecab.Destroy()

	lattice, err := NewLattice()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer lattice.Destroy()

	lattice.SetSentence("こんにちは世界")
	err = mecab.ParseLattice(lattice)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	expected := "こんにちは\t感動詞,*,*,*,*,*,こんにちは,コンニチハ,コンニチワ\n" +
		"世界\t名詞,一般,*,*,*,*,世界,セカイ,セカイ\n" +
		"EOS\n"
	if lattice.String() != expected {
		t.Errorf("expected %s, but %s", expected, lattice.String())
	}
}

func BenchmarkParseLattice(b *testing.B) {
	mecab, _ := New(rcfile(map[string]string{
		"output-format-type": "wakati",
	}))
	defer mecab.Destroy()

	lattice, _ := NewLattice()
	defer lattice.Destroy()

	for i := 0; i < b.N; i++ {
		lattice.SetSentence("こんにちは世界")
		mecab.ParseLattice(lattice)
		lattice.String()
	}
}

func TestParseToNode(t *testing.T) {
	mecab, err := New(rcfile(map[string]string{}))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer mecab.Destroy()

	// XXX: avoid GC, MeCab 0.996 has GC problem (see https://github.com/taku910/mecab/pull/24)
	mecab.Parse("")

	node, err := mecab.ParseToNode("こんにちは世界")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	node = node.Next()
	if node.Surface() != "こんにちは" {
		t.Errorf("want こんにちは, but %s", node.Surface())
	}
	node = node.Next()
	if node.Surface() != "世界" {
		t.Errorf("want 世界, but %s", node.Surface())
	}
}

func TestParseToWordNodes(t *testing.T) {
	mecab, err := New(rcfile(map[string]string{}))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer mecab.Destroy()

	// XXX: avoid GC, MeCab 0.996 has GC problem (see https://github.com/taku910/mecab/pull/24)
	mecab.Parse("")

	wordNodeList, err := mecab.ParseToWordNodes("こんにちは世界  こんばんは世界")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(wordNodeList) != 2 {
		t.Errorf("want 2, but %d", len(wordNodeList))
	}
	if len(wordNodeList[0].Nodes) != 2 {
		t.Errorf("want 2, but %d", len(wordNodeList[0].Nodes))
	}
	if len(wordNodeList[1].Nodes) != 2 {
		t.Errorf("want 2, but %d", len(wordNodeList[1].Nodes))
	}

	wordNodeStrings := fmt.Sprintf("%s|%s", wordNodeList[0].Word, wordNodeList[1].Word)
	if wordNodeStrings != "こんにちは世界|こんばんは世界" {
		t.Errorf("want こんにちは世界|こんばんは世界, but %s", wordNodeStrings)
	}

	surfaceStrings := fmt.Sprintf(
		"%s|%s|%s|%s",
		wordNodeList[0].Nodes[0].Surface,
		wordNodeList[0].Nodes[1].Surface,
		wordNodeList[1].Nodes[0].Surface,
		wordNodeList[1].Nodes[1].Surface,
	)
	if surfaceStrings != "こんにちは|世界|こんばんは|世界" {
		t.Errorf("want こんにちは|世界|こんばんは|世界, but %s", surfaceStrings)
	}
}

func TestMeCabFinalizer(t *testing.T) {
	for i := 0; i < 10000; i++ {
		New(rcfile(map[string]string{}))
	}
	runtime.GC()
	runtime.GC()
	runtime.GC()
}

func BenchmarkParseToNode(b *testing.B) {
	mecab, _ := New(rcfile(map[string]string{}))
	defer mecab.Destroy()

	// XXX: avoid GC, MeCab 0.996 has GC problem (see https://github.com/taku910/mecab/pull/24)
	mecab.Parse("")

	for i := 0; i < b.N; i++ {
		for node, _ := mecab.ParseToNode("こんにちは世界"); !node.IsZero(); node = node.Next() {
			node.Surface()
		}
	}
}
