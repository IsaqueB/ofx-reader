package sgml

import "strings"

type Node struct {
	Name     string
	Text     string
	Children []*Node
}

func (n *Node) Child(name string) *Node {
	if n == nil {
		return nil
	}
	name = strings.ToUpper(name)
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (n *Node) ChildrenNamed(name string) []*Node {
	if n == nil {
		return nil
	}
	name = strings.ToUpper(name)
	var out []*Node
	for _, c := range n.Children {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func (n *Node) Value(name string) string {
	c := n.Child(name)
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Text)
}

func (n *Node) Descendants(name string) []*Node {
	if n == nil {
		return nil
	}
	name = strings.ToUpper(name)
	var out []*Node
	var walk func(*Node)
	walk = func(cur *Node) {
		for _, c := range cur.Children {
			if c.Name == name {
				out = append(out, c)
			}
			walk(c)
		}
	}
	walk(n)
	return out
}
