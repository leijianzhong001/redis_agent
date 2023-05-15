package rax

type Rax struct {
	head     *RaxNode
	numele   uint64
	numnodes uint64
}

type RaxNode struct {
	isKey         bool
	isnull        bool
	iscompr       bool
	size          int32
	routeKey      []string
	childPointers []*RaxNode
	value         interface{}
}
