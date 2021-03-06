package firetest

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type node struct {
	value     interface{}
	children  map[string]*node
	parent    *node
	sliceKids bool
}

func newNode(data interface{}) *node {
	n := &node{children: map[string]*node{}}

	switch data := data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			child := newNode(v)
			child.parent = n
			n.children[k] = child
		}
	case map[string]string:
		for k, v := range data {
			child := newNode(v)
			child.parent = n
			n.children[k] = child
		}
	case []interface{}:
		n.sliceKids = true
		for i, v := range data {
			child := newNode(v)
			child.parent = n
			n.children[fmt.Sprint(i)] = child
		}
	case string, int, int8, int16, int32, int64, float32, float64, bool:
		n.value = data
	case nil:
		// do nothing
	default:
		panic(fmt.Sprintf("Type(%T) not supported\n", data))
	}

	return n
}

func (n *node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.objectify())
}

func (n *node) objectify() interface{} {
	if n.isNil() {
		return nil
	}

	if n.value != nil {
		return n.value
	}

	if n.sliceKids {
		obj := make([]interface{}, len(n.children))
		for k, v := range n.children {
			index, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			obj[index] = v.objectify()
		}
		return obj
	}

	obj := map[string]interface{}{}
	for k, v := range n.children {
		obj[k] = v.objectify()
	}

	return obj
}

func (n *node) isNil() bool {
	return n.value == nil && len(n.children) == 0
}

func (n *node) merge(newNode *node) {
	for k, v := range newNode.children {
		n.children[k] = v
	}
	n.value = newNode.value
}

func (n *node) prune() *node {
	if len(n.children) > 0 || n.value != nil {
		return nil
	}

	parent := n.parent
	n.parent = nil
	return parent
}
