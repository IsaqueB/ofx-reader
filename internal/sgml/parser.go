package sgml

import (
	"errors"
	"fmt"
	"html"
	"strings"
)

// Parse parses OFX 1.x SGML. It supports omitted end tags for leaf elements.
func Parse(input string) (*Node, error) {
	root := &Node{Name: "#DOCUMENT"}
	stack := []*Node{root}
	i := 0

	for i < len(input) {
		lt := strings.IndexByte(input[i:], '<')
		if lt < 0 {
			appendText(stack[len(stack)-1], input[i:])
			break
		}
		lt += i
		if lt > i {
			appendText(stack[len(stack)-1], input[i:lt])
		}
		gtRel := strings.IndexByte(input[lt:], '>')
		if gtRel < 0 {
			return nil, errors.New("ofx sgml: unterminated tag")
		}
		gt := lt + gtRel
		raw := strings.TrimSpace(input[lt+1 : gt])
		i = gt + 1
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "!") || strings.HasPrefix(raw, "?") {
			continue
		}

		// A leaf node in OFX 1.x commonly has text but no explicit closing tag.
		// When another tag begins, close that leaf implicitly.
		for len(stack) > 1 && strings.TrimSpace(stack[len(stack)-1].Text) != "" {
			stack = stack[:len(stack)-1]
		}

		if strings.HasPrefix(raw, "/") {
			name := tagName(raw[1:])
			found := -1
			for j := len(stack) - 1; j >= 1; j-- {
				if stack[j].Name == name {
					found = j
					break
				}
			}
			if found >= 0 {
				stack = stack[:found]
			}
			continue
		}

		name := tagName(raw)
		if name == "" {
			return nil, fmt.Errorf("ofx sgml: invalid tag <%s>", raw)
		}
		n := &Node{Name: name}
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, n)
		stack = append(stack, n)
	}

	if len(root.Children) == 0 {
		return nil, errors.New("ofx sgml: empty document")
	}
	return root, nil
}

func appendText(n *Node, s string) {
	if n == nil || s == "" {
		return
	}
	n.Text += html.UnescapeString(s)
}

func tagName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.IndexAny(raw, " \t\r\n/"); i >= 0 {
		raw = raw[:i]
	}
	return strings.ToUpper(raw)
}
