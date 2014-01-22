// Copyright 2014 Gyepi Sam. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redux

import (
	"math/rand"
	"strings"
	"testing"
)

// Create and test a maximally linked dependency graph with no cycles.
// Since most nodes have multiple dependencies and prerequisites, a node
// will be invoked as a redo-ifchange prerequisite multiple times.
// This test verifies that despite the multiple invocations, each node is built just once.
// The test also varies the dependency orderings to ensure that the result does not depend
// in some particularly auspicious order.
func TestDeepTree(t *testing.T) {
	const N = 10 // creates a tree with N nodes and N * (N - 1) / 2 dependencies
	const out = "1"

	tree := make([]Script, N)
	head := make([]string, N) //part of script prior to body
	tail := make([]string, N) //part of script after the body

	// A script that counts its invocations.
	// Multiple invocations will produce incorrect output (the number of extra invocations).
	// Each script's output will be included in its dependencies' output.
	body := `
value=1
if test -e $1 ; then
value=$(expr $(cat $1) + 1)
fi
printf "%d" $value
`
	for i := 0; i < N; i++ {
		name := string('A' + i)
		tree[i] = Script{Name: name, Out: out}
		head[i] = "redo-ifchange " + name
		tail[i] = "cat " + name
	}

	for _, order := range []string{"forward", "reverse", "shuffle"} {
		for k := 0; k < N; k++ {
			tmp := head[k+1 : N] //each node depends on all succeeding nodes.
			head0 := make([]string, len(tmp))
			copy(head0, tmp)

			tmp = tail[k+1 : N]
			tail0 := make([]string, len(tmp))
			copy(tail0, tmp)

			switch order {
			case "forward":
				//do nothing
			case "reverse":
				for i, j := 0, len(head0)-1; i < j; i, j = i+1, j-1 {
					head0[i], head0[j] = head0[j], head0[i]
					tail0[i], tail0[j] = tail0[j], tail0[i]
				}
			case "shuffle":
				//fischer yates shuffle
				for i := len(head0) - 1; i > 0; i-- {
					j := rand.Intn(i + 1)
					head0[i], head0[j] = head0[j], head0[i]
					tail0[i], tail0[j] = tail0[j], tail0[i]
				}
			default:
				panic("unknown order: " + order)
			}

			tree[k].Command = strings.Join(head0, "\n") + body + strings.Join(tail0, "\n")
		}

		// Each new node N+1 becomes a prerequisite for each of the previous N nodes.
		// Since each node, includes its prerequisite's output, the N+1'st node increases
		// the output by N. The total output size is 2^(N-1).
		// One could also use a counting argument: 1 node produces 1, 2 -> 2, 3 -> 4, 4 -> 8, etc.
		tree[0].Out = strings.Repeat(out, 1<<uint(len(tree)-1))
		t.Logf("DeepTree order: %s\n", order)
		SimpleTree(t, tree...)
	}
}

// redo-ifchange dependency on the same file, either repeated or via multiple paths,
// should resolve to a single dependency
func TestSimpleTree(t *testing.T) {
	s0 := Script{Name: "A", Out: "AB"}
	s0.Command = `
redo-ifchange B B ./B $(dirname $3)/B
echo -n A
cat B
`
	s1 := Script{Name: "B"}
	s1.Command = `
#Produce output for each invocation
if test -e $1 ; then
 cat $1
fi
echo -n B
`
	SimpleTree(t, s0, s1)
}
