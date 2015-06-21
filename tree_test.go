package firetest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNodeWithKids(children map[string]*node) *node {
	return &node{
		children: children,
	}
}

func equalNodes(expected, actual *node) error {
	if ec, ac := len(expected.children), len(actual.children); ec != ac {
		return fmt.Errorf("Children count is not the same\n\tExpected: %d\n\tActual: %d", ec, ac)
	}

	if len(expected.children) == 0 {
		if !assert.ObjectsAreEqualValues(expected.value, actual.value) {
			return fmt.Errorf("Node values not equal\n\tExpected: %T %v\n\tActual: %T %v", expected.value, expected.value, actual.value, actual.value)
		}
		return nil
	}

	for child, n := range expected.children {
		n2, ok := actual.children[child]
		if !ok {
			return fmt.Errorf("Expected node to have child: %s", child)
		}

		err := equalNodes(n, n2)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestNewNode(t *testing.T) {

	for _, test := range []struct {
		name string
		node *node
	}{
		{
			name: "scalars/string",
			node: newNode("foo"),
		},
		{
			name: "scalars/number",
			node: newNode(2),
		},
		{
			name: "scalars/decimal",
			node: newNode(2.2),
		},
		{
			name: "scalars/boolean",
			node: newNode(false),
		},
		{
			name: "arrays/strings",
			node: newNode([]interface{}{"foo", "bar"}),
		},
		{
			name: "arrays/booleans",
			node: newNode([]interface{}{true, false}),
		},
		{
			name: "arrays/numbers",
			node: newNode([]interface{}{1, 2, 3}),
		},
		{
			name: "arrays/decimals",
			node: newNode([]interface{}{1.1, 2.2, 3.3}),
		},
		{
			name: "objects/simple",
			node: newTestNodeWithKids(map[string]*node{
				"foo": newNode("bar"),
			}),
		},
		{
			name: "objects/complex",
			node: newTestNodeWithKids(map[string]*node{
				"foo":  newNode("bar"),
				"foo1": newNode(2),
				"foo2": newNode(true),
				"foo3": newNode(3.42),
			}),
		},
		{
			name: "objects/nested",
			node: newTestNodeWithKids(map[string]*node{
				"dinosaurs": newTestNodeWithKids(map[string]*node{
					"bruhathkayosaurus": newTestNodeWithKids(map[string]*node{
						"appeared": newNode(-70000000),
						"height":   newNode(25),
						"length":   newNode(44),
						"order":    newNode("saurischia"),
						"vanished": newNode(-70000000),
						"weight":   newNode(135000),
					}),
					"lambeosaurus": newTestNodeWithKids(map[string]*node{
						"appeared": newNode(-76000000),
						"height":   newNode(2.1),
						"length":   newNode(12.5),
						"order":    newNode("ornithischia"),
						"vanished": newNode(-75000000),
						"weight":   newNode(5000),
					}),
				}),
				"scores": newTestNodeWithKids(map[string]*node{
					"bruhathkayosaurus": newNode(55),
					"lambeosaurus":      newNode(21),
				}),
			}),
		},
		{
			name: "objects/with_arrays",
			node: newTestNodeWithKids(map[string]*node{
				"regular":  newNode("item"),
				"booleans": newNode([]interface{}{false, true}),
				"numbers":  newNode([]interface{}{1, 2}),
				"decimals": newNode([]interface{}{1.1, 2.2}),
				"strings":  newNode([]interface{}{"foo", "bar"}),
			}),
		},
	} {
		data, err := ioutil.ReadFile("fixtures/" + test.name + ".json")
		require.NoError(t, err, test.name)

		var v interface{}
		require.NoError(t, json.Unmarshal(data, &v), test.name)

		n := newNode(v)
		assert.NoError(t, equalNodes(test.node, n), test.name)
	}
}

func TestObjectify(t *testing.T) {
	for _, test := range []struct {
		name   string
		object interface{}
		node   *node
	}{
		{
			name:   "string",
			object: "foo",
		},
		{
			name:   "number",
			object: 2,
		},
		{
			name:   "decimal",
			object: 2.2,
		},
		{
			name:   "boolean",
			object: false,
		},
		{
			name:   "arrays",
			object: []interface{}{"foo", 2, 2.2, false},
		},
		{
			name: "object",
			object: map[string]interface{}{
				"one_fish":     "two_fish",
				"red_fish":     2.2,
				"netflix_list": []interface{}{"Orange is the New Black", "House of Cards"},
				"shopping_list": map[string]interface{}{
					"publix":  "milk",
					"walmart": "reese's pieces",
				},
			},
		},
	} {
		node := newNode(test.object)
		assert.Equal(t, test.object, node.objectify())
	}
}

func TestTreeAdd(t *testing.T) {
	for _, test := range []struct {
		path string
		node *node
	}{
		{
			path: "scalars/string",
			node: newNode("foo"),
		},
		{
			path: "s/c/a/l/a/r/s/s/t/r/i/n/g",
			node: newNode([]interface{}{"foo", "bar"}),
		},
	} {
		tree := newTree()
		tree.add(test.path, test.node)

		rabbitHole := strings.Split(test.path, "/")
		previous := tree.rootNode
		for i := 0; i < len(rabbitHole); i++ {
			var ok bool
			previous, ok = previous.children[rabbitHole[i]]
			require.True(t, ok, test.path)
		}

		assert.NoError(t, equalNodes(test.node, previous), test.path)
	}
}

func TestTreeGet(t *testing.T) {
	for _, test := range []struct {
		path string
		node *node
	}{
		{
			path: "scalars/string",
			node: newNode("foo"),
		},
		{
			path: "s/c/a/l/a/r/s/s/t/r/i/n/g",
			node: newNode([]interface{}{"foo", "bar"}),
		},
	} {
		tree := newTree()
		tree.add(test.path, test.node)

		assert.NoError(t, equalNodes(test.node, tree.get(test.path)), test.path)
	}
}
