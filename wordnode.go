package mecab

type WordNode struct {
	Word              string
	Nodes             []TokenizedWord
	WordLength        int
	NodeSurfaceLength int
}

type TokenizedWord struct {
	Surface string
	Feature string
}
