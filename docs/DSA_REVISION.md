# DSA Interview Revision (Go)

> **How to use this doc on interview day.** Read Section 1 (complexity) and Section 9 (pattern decision table) first — they prime your brain. Then skim the pattern triggers in bold ("**When you see…**"). Each pattern has a copy-paste Go template you can adapt in 60 seconds. Don't read top-to-bottom; jump via the table of contents.

## Table of Contents

**Part I — Core**
1. [Complexity Cheat Sheet](#1-complexity-cheat-sheet)
2. [Go DSA Toolkit (stdlib + idioms)](#2-go-dsa-toolkit)
3. [Core Data Structures](#3-core-data-structures)
4. [Two Pointers / Sliding Window / Fast-Slow](#4-two-pointers--sliding-window--fast-slow)
5. [Binary Search (all variants)](#5-binary-search)
6. [Trees](#6-trees)
7. [Graphs (BFS, DFS, Topo, Dijkstra, Union-Find)](#7-graphs)
8. [Recursion, Backtracking, Dynamic Programming, Greedy](#8-recursion-backtracking-dp-greedy)
9. [Pattern Decision Table — "What do I reach for?"](#9-pattern-decision-table)
10. [Sorting Algorithms](#10-sorting-algorithms)
11. [Bit Manipulation](#11-bit-manipulation)
12. [Math & Number Theory snippets](#12-math--number-theory)

**Part II — Advanced**
13. [Advanced Data Structures (Fenwick, Segment Tree, LRU, Sweep Line)](#13-advanced-data-structures)
14. [String Algorithms (KMP, Rabin-Karp, Z, Manacher)](#14-string-algorithms)
15. [Advanced Graph Algorithms (MST, Bellman-Ford, Floyd-Warshall, SCC, 0-1 BFS)](#15-advanced-graph-algorithms)
16. [Advanced DP Patterns (Interval, Bitmask, Tree, Digit DP)](#16-advanced-dp-patterns)

**Part III — Drill & Reference**
17. [Problem Drill & Worked Examples (~95 problems by pattern)](#17-problem-drill--worked-examples)
18. [Interview-Day Gotchas & Tips](#18-interview-day-gotchas--tips)

---

## 1. Complexity Cheat Sheet

### Big-O at a glance (what's acceptable for an input size N)

| N (input size) | Acceptable complexity | Typical pattern |
|----------------|----------------------|-----------------|
| N ≤ 10–12 | O(N!) | permutations / brute backtracking |
| N ≤ 20 | O(2^N) | subsets / bitmask DP |
| N ≤ 500 | O(N³) | DP on intervals, Floyd-Warshall |
| N ≤ 5,000 | O(N²) | nested loops, classic DP |
| N ≤ 10⁶ | O(N log N) | sort, heap, divide & conquer |
| N ≤ 10⁸ | O(N) / O(N log log N) | scan, sliding window, sieve |
| huge / queries | O(log N) / O(1) | binary search, hashing, math |

> Rule of thumb: ~10⁸ simple operations per second. If `N² > 10⁸`, you need a better-than-quadratic approach.

### Data structure operations

| Structure | Access | Search | Insert | Delete | Notes |
|-----------|:------:|:------:|:------:|:------:|-------|
| Array / Slice | O(1) | O(N) | O(N)* | O(N) | *append amortized O(1) |
| Hash Map / Set | — | O(1) avg | O(1) avg | O(1) avg | O(N) worst (collisions) |
| Linked List | O(N) | O(N) | O(1)† | O(1)† | †given the node ref |
| Stack / Queue | — | O(N) | O(1) | O(1) | LIFO / FIFO |
| Binary Heap | O(1) peek | O(N) | O(log N) | O(log N) | pop-min/max O(log N) |
| Balanced BST | O(log N) | O(log N) | O(log N) | O(log N) | ordered iteration |
| Trie | — | O(L) | O(L) | O(L) | L = key length |
| Union-Find | — | ~O(1) | ~O(1) | — | α(N) with path compression |

### Sorting

| Algorithm | Best | Average | Worst | Space | Stable |
|-----------|:----:|:-------:|:-----:|:-----:|:------:|
| Quicksort | N log N | N log N | N² | log N | No |
| Mergesort | N log N | N log N | N log N | N | Yes |
| Heapsort | N log N | N log N | N log N | 1 | No |
| Insertion | N | N² | N² | 1 | Yes |
| Counting | N+K | N+K | N+K | N+K | Yes |
| Radix | N·K | N·K | N·K | N+K | Yes |

`sort.Slice` (Go stdlib) is pattern-defeating quicksort, **not stable**. Use `sort.SliceStable` when ties must keep order.

### Recursion complexity (Master Theorem quick form)

For `T(N) = a·T(N/b) + O(N^d)`:
- `d > log_b(a)` → **O(N^d)** (work dominated by the split/combine)
- `d = log_b(a)` → **O(N^d · log N)** (e.g. mergesort: a=2,b=2,d=1 → N log N)
- `d < log_b(a)` → **O(N^(log_b a))** (leaves dominate)

---

## 2. Go DSA Toolkit

### Slices — the workhorse

```go
s := make([]int, 0, n)      // len 0, cap n (preallocate to avoid reallocs)
s = append(s, x)            // amortized O(1)
last := s[len(s)-1]         // peek
s = s[:len(s)-1]            // pop (stack)
s = s[1:]                   // pop front — O(1) but leaks underlying array head

// Delete index i (order-preserving, O(N)):
s = append(s[:i], s[i+1:]...)

// Delete index i (fast, O(1), order NOT preserved):
s[i] = s[len(s)-1]; s = s[:len(s)-1]

// Copy (slices are reference types — copy when you need independence):
dst := make([]int, len(src)); copy(dst, src)

// 2D grid:
grid := make([][]int, rows)
for i := range grid { grid[i] = make([]int, cols) }
```

> **Gotcha:** sub-slicing shares the backing array. `b := a[1:3]` then writing `b[0]` mutates `a[1]`. Reslicing within cap can also clobber data via `append`. Copy when in doubt.

### Maps & Sets

```go
m := map[string]int{}
m[k]++                       // zero value is 0, so this just works as a counter
v, ok := m[k]                // ok == false if absent
delete(m, k)

set := map[int]struct{}{}    // struct{} uses zero memory for the value
set[x] = struct{}{}
_, exists := set[x]
```

> Maps have **no defined iteration order** — never rely on it. To iterate sorted, collect keys into a slice and `sort` them.

### Heap (`container/heap`) — you implement the interface

```go
import "container/heap"

type MinHeap []int
func (h MinHeap) Len() int            { return len(h) }
func (h MinHeap) Less(i, j int) bool  { return h[i] < h[j] } // '>' for max-heap
func (h MinHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x any)         { *h = append(*h, x.(int)) }
func (h *MinHeap) Pop() any {
    old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x
}

// Usage:
h := &MinHeap{5, 2, 8}
heap.Init(h)
heap.Push(h, 1)
min := heap.Pop(h).(int)   // smallest out first
```

> For a **priority queue of structs**, store `[]Item` and compare a `priority` field in `Less`. For max-heap, flip the comparison or push negative values.

### Sorting

```go
sort.Ints(a); sort.Strings(b); sort.Float64s(c)
sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })      // unstable
sort.SliceStable(a, func(i, j int) bool { return a[i] < a[j] })
idx := sort.SearchInts(a, target)   // leftmost index where target could insert
```

### Other stdlib worth knowing

```go
import "math"
math.MaxInt, math.MinInt, math.MaxInt32, math.Inf(1)
import "strings"
strings.Builder{}            // O(N) string concatenation (don't += in loops)
import "container/list"      // doubly linked list (rarely needed; slices usually win)
```

### Common Go idioms for DSA

```go
// abs (no built-in for ints):
func abs(x int) int { if x < 0 { return -x }; return x }
func max(a, b int) int { if a > b { return a }; return b }  // Go 1.21+ has builtin max/min
func min(a, b int) int { if a < b { return a }; return b }

// Multi-return for clarity in BFS/DFS coords:
type pair struct{ r, c int }
```

---

## 3. Core Data Structures

### 3.1 Linked List

```go
type ListNode struct {
    Val  int
    Next *ListNode
}
```

**Reverse a linked list** (the single most common LL operation — memorize it):

```go
func reverse(head *ListNode) *ListNode {
    var prev *ListNode
    for head != nil {
        next := head.Next   // save
        head.Next = prev    // flip
        prev = head         // advance prev
        head = next         // advance head
    }
    return prev
}
```

**Dummy head trick** — eliminates edge cases when the head itself might change:

```go
dummy := &ListNode{Next: head}
prev := dummy
// ... manipulate, then:
return dummy.Next
```

**When you see** "remove Nth from end", "merge two lists", "detect/remove cycle", "reorder" → linked list with **two pointers** and a **dummy head**.

### 3.2 Stack & Queue (just use a slice)

```go
// Stack (LIFO)
stack := []int{}
stack = append(stack, x)            // push
top := stack[len(stack)-1]          // peek
stack = stack[:len(stack)-1]        // pop

// Queue (FIFO) — slice as queue
queue := []int{x}
front := queue[0]                   // peek
queue = queue[1:]                   // dequeue
```

**When you see** "matching parentheses", "next greater element", "evaluate expression", "undo", "nested structure" → **stack**. "Level-by-level", "shortest path in unweighted graph", "process in arrival order" → **queue**.

### 3.3 Monotonic Stack (huge for "next greater/smaller")

Keeps elements in increasing or decreasing order; each element pushed/popped once → **O(N)**.

```go
// Next Greater Element to the right. res[i] = index of next greater, else -1.
func nextGreater(nums []int) []int {
    n := len(nums)
    res := make([]int, n)
    for i := range res { res[i] = -1 }
    stack := []int{} // holds indices, values decreasing
    for i := 0; i < n; i++ {
        for len(stack) > 0 && nums[i] > nums[stack[len(stack)-1]] {
            top := stack[len(stack)-1]
            stack = stack[:len(stack)-1]
            res[top] = i
        }
        stack = append(stack, i)
    }
    return res
}
```

**When you see** "next greater/smaller", "largest rectangle in histogram", "stock span", "daily temperatures", "trapping rain water" → **monotonic stack**.

### 3.4 Trie (prefix tree)

```go
type Trie struct {
    children [26]*Trie
    isEnd    bool
}

func (t *Trie) Insert(word string) {
    node := t
    for i := 0; i < len(word); i++ {
        c := word[i] - 'a'
        if node.children[c] == nil {
            node.children[c] = &Trie{}
        }
        node = node.children[c]
    }
    node.isEnd = true
}

func (t *Trie) Search(word string) bool {
    node := t.find(word)
    return node != nil && node.isEnd
}
func (t *Trie) StartsWith(prefix string) bool { return t.find(prefix) != nil }

func (t *Trie) find(s string) *Trie {
    node := t
    for i := 0; i < len(s); i++ {
        c := s[i] - 'a'
        if node.children[c] == nil { return nil }
        node = node.children[c]
    }
    return node
}
```

**When you see** "autocomplete", "prefix", "dictionary of words", "word search II" → **trie**.

### 3.5 Heap / Priority Queue patterns

- **Top K elements** → keep a min-heap of size K (push, pop when size > K). O(N log K).
- **Kth largest** → min-heap of size K; root is the answer.
- **Merge K sorted lists** → heap of the K current heads.
- **Median of a stream** → two heaps (max-heap for low half, min-heap for high half).
- **Dijkstra** → min-heap by distance.

---

## 4. Two Pointers / Sliding Window / Fast-Slow

### 4.1 Two Pointers (sorted array, pair-finding)

```go
// Two-sum on a SORTED array → target
func twoSumSorted(a []int, target int) (int, int) {
    l, r := 0, len(a)-1
    for l < r {
        sum := a[l] + a[r]
        switch {
        case sum == target: return l, r
        case sum < target:  l++
        default:            r--
        }
    }
    return -1, -1
}
```

**When you see** "sorted array + find pair/triple", "remove duplicates in place", "container with most water", "3sum", "palindrome check" → **two pointers**.

### 4.2 Sliding Window (subarray / substring)

The single most reused interview pattern. Two flavors:

**Fixed-size window:**
```go
func maxSumWindow(a []int, k int) int {
    sum := 0
    for i := 0; i < k; i++ { sum += a[i] }
    best := sum
    for i := k; i < len(a); i++ {
        sum += a[i] - a[i-k]    // slide: add new, drop old
        best = max(best, sum)
    }
    return best
}
```

**Variable-size window (expand right, shrink left while invalid):**
```go
// Longest substring without repeating characters
func lengthOfLongestSubstring(s string) int {
    seen := map[byte]int{} // char -> last index
    best, left := 0, 0
    for right := 0; right < len(s); right++ {
        c := s[right]
        if idx, ok := seen[c]; ok && idx >= left {
            left = idx + 1          // jump left past the duplicate
        }
        seen[c] = right
        best = max(best, right-left+1)
    }
    return best
}
```

**Template for "minimum window / at most K / exactly K":**
```go
left := 0
for right := 0; right < n; right++ {
    // 1. include a[right] into window state
    for windowIsInvalid() {
        // 2. remove a[left] from window state
        left++
    }
    // 3. window [left..right] is now valid → update answer
}
```

> Trick: "**exactly K**" distinct = "at most K" − "at most K−1".

**When you see** "longest/shortest/max/min subarray or substring", "contiguous", "at most/exactly K" → **sliding window**.

### 4.3 Fast & Slow Pointers (cycle detection)

```go
// Floyd's cycle detection in a linked list
func hasCycle(head *ListNode) bool {
    slow, fast := head, head
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
        if slow == fast { return true }
    }
    return false
}
```

**When you see** "cycle", "find middle of list", "happy number", "find duplicate number (array as implicit linked list)" → **fast/slow pointers**.

---

## 5. Binary Search

The bug magnet. Memorize **one** correct template and adapt it. Use the half-open `[lo, hi)` form:

```go
// Leftmost index where pred(i) is true; pred must be monotonic: F F F T T T
func binarySearch(n int, pred func(int) bool) int {
    lo, hi := 0, n          // hi is exclusive
    for lo < hi {
        mid := lo + (hi-lo)/2   // avoids overflow
        if pred(mid) {
            hi = mid            // answer is mid or to the left
        } else {
            lo = mid + 1        // answer is strictly right
        }
    }
    return lo               // == n if no index satisfies pred
}
```

**Classic value search:**
```go
func search(a []int, target int) int {
    lo, hi := 0, len(a)-1
    for lo <= hi {
        mid := lo + (hi-lo)/2
        switch {
        case a[mid] == target: return mid
        case a[mid] < target:  lo = mid + 1
        default:               hi = mid - 1
        }
    }
    return -1
}
```

**Binary search on the ANSWER** (the powerful, under-recognized variant): when the question asks for a min/max value and you can write a `feasible(x) bool` that's monotonic, binary search over the value range.

```go
// "Minimum eating speed / capacity / days such that condition holds"
func minFeasible(lo, hi int, feasible func(int) bool) int {
    for lo < hi {
        mid := lo + (hi-lo)/2
        if feasible(mid) { hi = mid } else { lo = mid + 1 }
    }
    return lo
}
```

**When you see** "sorted", "find min/max X such that…", "rotated sorted array", "smallest value that works", "Koko eating bananas", "split array largest sum", "capacity to ship" → **binary search** (often on the answer).

> **Off-by-one survival kit:** decide if `hi` is inclusive (`len-1`, loop `lo<=hi`) or exclusive (`len`, loop `lo<hi`) and stay consistent. `mid = lo + (hi-lo)/2` always. When shrinking, one side moves past mid (`±1`), the other keeps mid.

---

## 6. Trees

```go
type TreeNode struct {
    Val         int
    Left, Right *TreeNode
}
```

### 6.1 Traversals (DFS)

```go
// Inorder (Left, Root, Right) — gives SORTED order for a BST
func inorder(root *TreeNode, out *[]int) {
    if root == nil { return }
    inorder(root.Left, out)
    *out = append(*out, root.Val)
    inorder(root.Right, out)
}
// Preorder: Root, Left, Right   (serialize, copy tree)
// Postorder: Left, Right, Root  (delete tree, compute height bottom-up)
```

**Iterative inorder** (when recursion depth is a concern / asked explicitly):
```go
func inorderIter(root *TreeNode) []int {
    res, stack := []int{}, []*TreeNode{}
    cur := root
    for cur != nil || len(stack) > 0 {
        for cur != nil { stack = append(stack, cur); cur = cur.Left }
        cur = stack[len(stack)-1]; stack = stack[:len(stack)-1]
        res = append(res, cur.Val)
        cur = cur.Right
    }
    return res
}
```

### 6.2 BFS / Level-order

```go
func levelOrder(root *TreeNode) [][]int {
    if root == nil { return nil }
    res := [][]int{}
    queue := []*TreeNode{root}
    for len(queue) > 0 {
        n := len(queue)            // freeze this level's size
        level := make([]int, 0, n)
        for i := 0; i < n; i++ {
            node := queue[0]; queue = queue[1:]
            level = append(level, node.Val)
            if node.Left != nil  { queue = append(queue, node.Left) }
            if node.Right != nil { queue = append(queue, node.Right) }
        }
        res = append(res, level)
    }
    return res
}
```

### 6.3 Common tree recipes

```go
// Max depth
func maxDepth(r *TreeNode) int {
    if r == nil { return 0 }
    return 1 + max(maxDepth(r.Left), maxDepth(r.Right))
}

// Validate BST — pass down allowed (min, max) range
func isValidBST(r *TreeNode) bool {
    var rec func(n *TreeNode, lo, hi int) bool
    rec = func(n *TreeNode, lo, hi int) bool {
        if n == nil { return true }
        if n.Val <= lo || n.Val >= hi { return false }
        return rec(n.Left, lo, n.Val) && rec(n.Right, n.Val, hi)
    }
    return rec(r, math.MinInt, math.MaxInt)
}

// Lowest Common Ancestor (general binary tree)
func lca(root, p, q *TreeNode) *TreeNode {
    if root == nil || root == p || root == q { return root }
    l := lca(root.Left, p, q)
    r := lca(root.Right, p, q)
    if l != nil && r != nil { return root } // p and q split here
    if l != nil { return l }
    return r
}
```

**When you see** "path", "depth/height", "ancestor", "subtree", "balanced", "diameter" → **DFS recursion, return info up**. "Level", "width", "min depth", "nearest" → **BFS**. "BST + kth / range / sorted" → **inorder**.

---

## 7. Graphs

### 7.1 Representations

```go
// Adjacency list (most common). For weighted, store {to, w}.
adj := make([][]int, n)
adj[u] = append(adj[u], v)
adj[v] = append(adj[v], u)   // undirected: add both directions
```

### 7.2 BFS (shortest path in UNWEIGHTED graph)

```go
func bfs(adj [][]int, start int) []int {
    n := len(adj)
    dist := make([]int, n)
    for i := range dist { dist[i] = -1 }
    dist[start] = 0
    queue := []int{start}
    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
        for _, v := range adj[u] {
            if dist[v] == -1 {        // unvisited
                dist[v] = dist[u] + 1
                queue = append(queue, v)
            }
        }
    }
    return dist
}
```

### 7.3 DFS (connectivity, components, cycles)

```go
func dfs(adj [][]int, u int, visited []bool) {
    visited[u] = true
    for _, v := range adj[u] {
        if !visited[v] { dfs(adj, v, visited) }
    }
}
```

### 7.4 Grid BFS/DFS (super common — "islands", "rotting oranges", "maze")

```go
var dirs = [4][2]int{{1,0},{-1,0},{0,1},{0,-1}}

func numIslands(grid [][]byte) int {
    rows, cols := len(grid), len(grid[0])
    count := 0
    var sink func(r, c int)
    sink = func(r, c int) {
        if r < 0 || r >= rows || c < 0 || c >= cols || grid[r][c] != '1' { return }
        grid[r][c] = '0'                // mark visited in place
        for _, d := range dirs { sink(r+d[0], c+d[1]) }
    }
    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            if grid[r][c] == '1' { count++; sink(r, c) }
        }
    }
    return count
}
```

> **Multi-source BFS** ("rotting oranges", "walls and gates", "nearest 0"): seed the queue with ALL sources at distance 0, then BFS once.

### 7.5 Topological Sort (DAG ordering — "course schedule", "build order")

**Kahn's algorithm (BFS, also detects cycles):**
```go
func topoSort(n int, adj [][]int) ([]int, bool) {
    indeg := make([]int, n)
    for u := 0; u < n; u++ {
        for _, v := range adj[u] { indeg[v]++ }
    }
    queue := []int{}
    for i := 0; i < n; i++ { if indeg[i] == 0 { queue = append(queue, i) } }
    order := []int{}
    for len(queue) > 0 {
        u := queue[0]; queue = queue[1:]
        order = append(order, u)
        for _, v := range adj[u] {
            indeg[v]--
            if indeg[v] == 0 { queue = append(queue, v) }
        }
    }
    return order, len(order) == n   // false → graph has a cycle
}
```

**When you see** "prerequisites", "ordering with dependencies", "can you finish", "build/compile order" → **topological sort** (and cycle check).

### 7.6 Dijkstra (shortest path, NON-negative weights)

```go
import "container/heap"

type edge struct{ to, w int }
type item struct{ node, dist int }
type pq []item
func (p pq) Len() int            { return len(p) }
func (p pq) Less(i, j int) bool  { return p[i].dist < p[j].dist }
func (p pq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x any)         { *p = append(*p, x.(item)) }
func (p *pq) Pop() any           { o := *p; n := len(o); x := o[n-1]; *p = o[:n-1]; return x }

func dijkstra(adj [][]edge, src int) []int {
    n := len(adj)
    dist := make([]int, n)
    for i := range dist { dist[i] = math.MaxInt }
    dist[src] = 0
    h := &pq{{src, 0}}
    for h.Len() > 0 {
        cur := heap.Pop(h).(item)
        if cur.dist > dist[cur.node] { continue }   // stale entry — skip
        for _, e := range adj[cur.node] {
            if nd := cur.dist + e.w; nd < dist[e.to] {
                dist[e.to] = nd
                heap.Push(h, item{e.to, nd})
            }
        }
    }
    return dist
}
```

> Negative weights → **Bellman-Ford** (O(V·E)). All-pairs shortest → **Floyd-Warshall** (O(V³), simple triple loop).

### 7.7 Union-Find / Disjoint Set Union (DSU)

```go
type DSU struct{ parent, rank []int }

func NewDSU(n int) *DSU {
    p := make([]int, n)
    for i := range p { p[i] = i }
    return &DSU{p, make([]int, n)}
}
func (d *DSU) Find(x int) int {
    for d.parent[x] != x {
        d.parent[x] = d.parent[d.parent[x]] // path compression (halving)
        x = d.parent[x]
    }
    return x
}
func (d *DSU) Union(a, b int) bool {
    ra, rb := d.Find(a), d.Find(b)
    if ra == rb { return false }            // already connected
    if d.rank[ra] < d.rank[rb] { ra, rb = rb, ra }
    d.parent[rb] = ra
    if d.rank[ra] == d.rank[rb] { d.rank[ra]++ }
    return true
}
```

**When you see** "connected components", "redundant connection", "number of provinces", "accounts merge", "Kruskal's MST", "are X and Y in the same group" → **Union-Find**.

---

## 8. Recursion, Backtracking, DP, Greedy

### 8.1 Backtracking template (subsets, permutations, combinations, N-queens)

```go
func subsets(nums []int) [][]int {
    res := [][]int{}
    path := []int{}
    var backtrack func(start int)
    backtrack = func(start int) {
        // record current path (make a COPY — path is mutated in place)
        cp := make([]int, len(path)); copy(cp, path)
        res = append(res, cp)
        for i := start; i < len(nums); i++ {
            path = append(path, nums[i])   // choose
            backtrack(i + 1)               // explore
            path = path[:len(path)-1]      // un-choose (backtrack)
        }
    }
    backtrack(0)
    return res
}
```

The three lines **choose → explore → un-choose** are the heartbeat of every backtracking problem.

- **Permutations**: track a `used []bool`, iterate from 0 each call.
- **Combinations**: pass `start` so you never reuse earlier elements.
- **Avoid dup results** with a sorted input + `if i > start && nums[i] == nums[i-1] { continue }`.
- **Prune** aggressively (e.g. if partial sum already exceeds target, return) — this is what makes exponential search pass.

**When you see** "all combinations / permutations / subsets", "generate all", "partition", "place items under constraints", "sudoku/N-queens" → **backtracking**.

### 8.2 Dynamic Programming — the recognition + recipe

**You probably need DP when:** the problem asks for an **optimum** (max/min/longest/count of ways) AND choices at each step depend on earlier choices, AND brute force has overlapping subproblems.

**Recipe (always do these 4 steps out loud):**
1. **State**: what do `dp[i]` / `dp[i][j]` *mean*? (Define in one sentence.)
2. **Transition**: how does a state derive from smaller states?
3. **Base case**: smallest states' values.
4. **Order / answer**: iterate so dependencies are ready; identify which cell is the answer.

**1D DP — climbing stairs / house robber:**
```go
// House robber: max sum, no two adjacent
func rob(nums []int) int {
    prev, cur := 0, 0   // prev = best up to i-2, cur = best up to i-1
    for _, x := range nums {
        prev, cur = cur, max(cur, prev+x)
    }
    return cur
}
```

**0/1 Knapsack (capacity W, weights/values):**
```go
func knapsack(w, v []int, W int) int {
    dp := make([]int, W+1)              // dp[c] = best value with capacity c
    for i := range w {
        for c := W; c >= w[i]; c-- {    // iterate capacity DOWN for 0/1 use
            dp[c] = max(dp[c], dp[c-w[i]]+v[i])
        }
    }
    return dp[W]
}
```
> Unbounded knapsack (reuse items, "coin change") → iterate capacity **UP**.

**Coin Change (min coins to make amount):**
```go
func coinChange(coins []int, amount int) int {
    dp := make([]int, amount+1)
    for i := 1; i <= amount; i++ {
        dp[i] = amount + 1                 // "infinity"
        for _, c := range coins {
            if c <= i && dp[i-c]+1 < dp[i] { dp[i] = dp[i-c] + 1 }
        }
    }
    if dp[amount] > amount { return -1 }
    return dp[amount]
}
```

**Longest Increasing Subsequence (O(N log N) version):**
```go
func lengthOfLIS(nums []int) int {
    tails := []int{}     // tails[k] = smallest tail of an increasing subseq of length k+1
    for _, x := range nums {
        i := sort.SearchInts(tails, x)     // leftmost ≥ x
        if i == len(tails) {
            tails = append(tails, x)
        } else {
            tails[i] = x
        }
    }
    return len(tails)
}
```

**2D grid DP — unique paths / min path sum / edit distance / LCS:**
```go
// Longest Common Subsequence
func lcs(a, b string) int {
    m, n := len(a), len(b)
    dp := make([][]int, m+1)
    for i := range dp { dp[i] = make([]int, n+1) }
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if a[i-1] == b[j-1] {
                dp[i][j] = dp[i-1][j-1] + 1
            } else {
                dp[i][j] = max(dp[i-1][j], dp[i][j-1])
            }
        }
    }
    return dp[m][n]
}
```

**Common DP families to recognize:**

| Family | Signal | State idea |
|--------|--------|-----------|
| Linear / sequence | "ways to reach", "rob houses" | dp[i] from dp[i-1], dp[i-2] |
| Knapsack | "subset with sum / max value under capacity" | dp[i][capacity] |
| Grid | "paths in matrix", "min cost path" | dp[r][c] from top/left |
| Interval | "burst balloons", "matrix chain", "palindrome partition" | dp[i][j] for range i..j |
| Subsequence | LCS, LIS, edit distance | dp[i][j] over two indices |
| Bitmask | N ≤ 20, "visit all", TSP | dp[mask][i] |
| DP on trees | "rob in tree", subtree aggregates | return tuple up via DFS |

> **Memoization shortcut:** if defining iteration order is hard, write the recursion first, then add a `memo map[stateKey]int` (or 2D slice initialized to a sentinel) at the top. Top-down is often easier to get right under pressure.

### 8.3 Greedy

Make the locally optimal choice and never reconsider. **Only valid when a local optimum provably leads to a global optimum** — prove it (or argue an exchange argument) before trusting greedy.

Common greedy wins: interval scheduling (sort by end time), jump game (track furthest reach), gas station, Huffman, assigning tasks, "minimum number of arrows/platforms".

```go
// Interval scheduling: max non-overlapping intervals → sort by END, greedily take
func maxNonOverlap(intervals [][]int) int {
    sort.Slice(intervals, func(i, j int) bool { return intervals[i][1] < intervals[j][1] })
    count, end := 0, math.MinInt
    for _, iv := range intervals {
        if iv[0] >= end { count++; end = iv[1] }   // take it
    }
    return count
}
```

> **Intervals in general:** almost always start by **sorting** (by start OR by end — decide which). Merge intervals = sort by start, extend or push.

---

## 9. Pattern Decision Table

Read the **left column out of the problem statement** → reach for the right.

| If the problem says / has… | Reach for |
|-----------------------------|-----------|
| Sorted array, find pair/triple | Two pointers |
| Contiguous subarray/substring, longest/shortest/at-most-K | Sliding window |
| "Top K", "Kth largest", "merge K", streaming median | Heap |
| Next greater/smaller, histogram, span | Monotonic stack |
| Matching/nested/balanced, evaluate expression | Stack |
| Find min/max value satisfying a monotonic condition | Binary search on answer |
| Sorted / rotated sorted lookup | Binary search |
| Cycle, middle of list, duplicate-as-pointer | Fast & slow pointers |
| Tree path/height/ancestor/diameter | DFS, return info up |
| Tree levels / nearest / min depth | BFS |
| Grid islands/regions/flood | Grid DFS/BFS |
| Shortest path, unweighted | BFS |
| Shortest path, weighted non-negative | Dijkstra |
| Shortest path, negative weights | Bellman-Ford |
| Dependencies / prerequisites / ordering | Topological sort |
| Connected components / "same group" / MST | Union-Find |
| Prefix, autocomplete, dictionary | Trie |
| All subsets/permutations/combinations | Backtracking |
| Optimum (max/min/count) w/ overlapping subproblems | DP |
| N ≤ 20 and "visit all / assign" | Bitmask DP |
| Locally optimal choice provably global | Greedy |
| Range queries with updates | Segment tree / BIT (Fenwick) |
| Running sum / range sum (no updates) | Prefix sum |

### Prefix sum (cheap but frequently the trick)

```go
// Subarray sum equals K — count subarrays with sum == k in O(N)
func subarraySum(nums []int, k int) int {
    count, sum := 0, 0
    seen := map[int]int{0: 1}     // prefix-sum -> frequency
    for _, x := range nums {
        sum += x
        count += seen[sum-k]      // how many earlier prefixes make sum-k
        seen[sum]++
    }
    return count
}
```

---

## 10. Sorting Algorithms

Know how each works conceptually; in interviews you'll usually call `sort.Slice`, but you may be asked to implement or reason about one.

**Quicksort (partition idea):**
```go
func quicksort(a []int, lo, hi int) {
    if lo >= hi { return }
    pivot := a[hi]
    i := lo
    for j := lo; j < hi; j++ {
        if a[j] < pivot { a[i], a[j] = a[j], a[i]; i++ }
    }
    a[i], a[hi] = a[hi], a[i]   // pivot to its sorted place
    quicksort(a, lo, i-1)
    quicksort(a, i+1, hi)
}
```

**Mergesort (merge idea):** split in half, recursively sort, merge two sorted halves with two pointers. Stable, guaranteed O(N log N), O(N) extra space. The "merge two sorted arrays" subroutine appears on its own constantly.

**Counting / Radix:** when values are small integers in a bounded range → linear time. Mention these when an interviewer pushes "can you beat N log N?" and the keys are bounded.

---

## 11. Bit Manipulation

```go
x & 1                 // lowest bit (odd/even)
x >> k, x << k        // divide/multiply by 2^k
x & (x - 1)           // clears the lowest set bit
x & (-x)              // isolates the lowest set bit
x | (1 << i)          // set bit i
x & ^(1 << i)         // clear bit i
x ^ (1 << i)          // toggle bit i
(x >> i) & 1          // read bit i
bits.OnesCount(uint(x))   // popcount (import "math/bits")
```

**Idioms:**
- **XOR of all** finds the single non-duplicated number (`a ^ a == 0`).
- **XOR swap**, **XOR to find missing number** (xor 0..n with array).
- **Subset enumeration**: `for sub := mask; sub > 0; sub = (sub-1) & mask` iterates all submasks of `mask`.
- **Check power of two**: `x > 0 && x&(x-1) == 0`.

**When you see** "single number", "count bits", "without using + operator", "subsets via bitmask", "N ≤ 20 visit-all" → **bit tricks**.

---

## 12. Math & Number Theory

```go
// GCD (Euclid) — LCM via a/gcd*b
func gcd(a, b int) int { for b != 0 { a, b = b, a%b }; return a }
func lcm(a, b int) int { return a / gcd(a, b) * b }

// Sieve of Eratosthenes — all primes < n in O(n log log n)
func sieve(n int) []bool {
    isComposite := make([]bool, n)
    for i := 2; i*i < n; i++ {
        if !isComposite[i] {
            for j := i * i; j < n; j += i { isComposite[j] = true }
        }
    }
    return isComposite      // isComposite[p] == false (and p>=2) means prime
}

// Fast modular exponentiation: base^exp % mod in O(log exp)
func modpow(base, exp, mod int) int {
    res := 1 % mod
    base %= mod
    for exp > 0 {
        if exp&1 == 1 { res = res * base % mod }
        base = base * base % mod
        exp >>= 1
    }
    return res
}
```

- **Overflow**: Go `int` is 64-bit on modern platforms, but products of two large ints still overflow. Take `% mod` as you go; problems with big counts usually want the answer mod `1e9+7`.
- **Combinatorics**: nCr, Pascal's triangle DP, stars-and-bars for "ways to distribute".

---

## 13. Advanced Data Structures

> These show up in mid/senior interviews and competitive rounds. Each has a sharp trigger — learn the trigger, then the template. All Go below compiles on 1.21+ (builtin `min`/`max`).

### Quick chooser

| You need… | Reach for | Update | Query |
|-----------|-----------|:------:|:-----:|
| Prefix sums + point updates | Fenwick / BIT | O(log N) | O(log N) |
| Range query + point update | Segment Tree | O(log N) | O(log N) |
| Range update + range query | Lazy Segment Tree | O(log N) | O(log N) |
| O(1) get/put with eviction | LRU (map + DLL) | O(1) | O(1) |
| Peak concurrency / overlaps | Sweep line + heap | — | O(N log N) total |
| Stack min in O(1) | Min-Stack | O(1) | O(1) |

---

### 1. Fenwick Tree / Binary Indexed Tree

One-sentence: a compact array that does **point update + prefix-sum query** in O(log N) using the low-bit trick `i & -i`; reach for it over a segment tree when you only need prefix/range *sums*.

**When you see…** "count of smaller elements to the right", "running prefix sums with updates", "number of inversions", "range sum + point update", "cumulative frequency" → **Fenwick / BIT**.

```go
// Fenwick Tree / BIT: 1-indexed internally.
type Fenwick struct{ t []int }

func NewFenwick(n int) *Fenwick { return &Fenwick{t: make([]int, n+1)} }

// Add delta at index i (0-based).
func (f *Fenwick) Update(i, delta int) {
	for i++; i < len(f.t); i += i & (-i) {
		f.t[i] += delta
	}
}

// Prefix sum of [0..i] (0-based, inclusive).
func (f *Fenwick) PrefixSum(i int) int {
	s := 0
	for i++; i > 0; i -= i & (-i) {
		s += f.t[i]
	}
	return s
}

// Range sum [l..r] inclusive.
func (f *Fenwick) RangeSum(l, r int) int { return f.PrefixSum(r) - f.PrefixSum(l-1) }
```

**Complexity:** update O(log N), query O(log N), space O(N). Build = N updates → O(N log N), or O(N) with an in-place variant.

---

### 2. Segment Tree (range query + point update)

One-sentence: a binary tree over array ranges giving **any associative range query** (sum, min, max, gcd…) with point updates in O(log N); this is the iterative bottom-up form (leaves at `[n, 2n)`).

**When you see…** "range min/max/sum with updates", "range GCD", "queries interleaved with modifications", anything Fenwick can't do (min/max) → **Segment Tree**.

```go
// SegTree for range-sum + point-update. For range-min, swap '+' for min in
// merge() and use a neutral identity (math.MaxInt) instead of 0.
type SegTree struct {
	n    int
	tree []int
}

func merge(a, b int) int { return a + b } // for min: return min(a, b)

func NewSegTree(nums []int) *SegTree {
	n := len(nums)
	st := &SegTree{n: n, tree: make([]int, 2*n)}
	for i, v := range nums { // leaves live at [n .. 2n)
		st.tree[n+i] = v
	}
	for i := n - 1; i > 0; i-- { // internal nodes bottom-up
		st.tree[i] = merge(st.tree[2*i], st.tree[2*i+1])
	}
	return st
}

// Point update: set index i (0-based) to val.
func (st *SegTree) Update(i, val int) {
	i += st.n
	st.tree[i] = val
	for i > 1 {
		i /= 2
		st.tree[i] = merge(st.tree[2*i], st.tree[2*i+1])
	}
}

// Range query [l, r) — half-open, 0-based.
func (st *SegTree) Query(l, r int) int {
	res := 0 // identity for sum; use math.MaxInt for min
	for l, r = l+st.n, r+st.n; l < r; l, r = l/2, r/2 {
		if l&1 == 1 {
			res = merge(res, st.tree[l])
			l++
		}
		if r&1 == 1 {
			r--
			res = merge(res, st.tree[r])
		}
	}
	return res
}
```

> **Range-min variant:** change `merge` to `return min(a, b)` and set `res := math.MaxInt`. Same structure, different identity + combiner.

**Complexity:** build O(N), update O(log N), query O(log N), space O(N) (2N array, no recursion).

---

### 3. Lazy-Propagation Segment Tree (advanced — keep in back pocket)

One-sentence: a recursive segment tree that defers range updates as "lazy" tags pushed down only when needed, enabling **range update + range query** both in O(log N). Only reach for this when **both** the update *and* the query span ranges — otherwise the plain segment tree is simpler.

**When you see…** "add X to all elements in [l, r] AND query a range" repeatedly → **lazy segment tree**.

```go
// Lazy-propagation segment tree: range-add update + range-sum query.
// ADVANCED — only when BOTH update and query are over ranges.
// Root = node 1 covering [0, n-1]. Call with Update(1,0,n-1,...) / Query(1,0,n-1,...).
type LazySeg struct {
	n         int
	sum, lazy []int
}

func NewLazySeg(n int) *LazySeg {
	return &LazySeg{n: n, sum: make([]int, 4*n), lazy: make([]int, 4*n)}
}

func (s *LazySeg) push(node, lo, hi int) {
	if s.lazy[node] == 0 {
		return
	}
	mid := (lo + hi) / 2
	l, r := 2*node, 2*node+1
	s.sum[l] += s.lazy[node] * (mid - lo + 1)
	s.sum[r] += s.lazy[node] * (hi - mid)
	s.lazy[l] += s.lazy[node]
	s.lazy[r] += s.lazy[node]
	s.lazy[node] = 0
}

func (s *LazySeg) Update(node, lo, hi, ql, qr, val int) {
	if qr < lo || hi < ql {
		return
	}
	if ql <= lo && hi <= qr {
		s.sum[node] += val * (hi - lo + 1)
		s.lazy[node] += val
		return
	}
	s.push(node, lo, hi)
	mid := (lo + hi) / 2
	s.Update(2*node, lo, mid, ql, qr, val)
	s.Update(2*node+1, mid+1, hi, ql, qr, val)
	s.sum[node] = s.sum[2*node] + s.sum[2*node+1]
}

func (s *LazySeg) Query(node, lo, hi, ql, qr int) int {
	if qr < lo || hi < ql {
		return 0
	}
	if ql <= lo && hi <= qr {
		return s.sum[node]
	}
	s.push(node, lo, hi)
	mid := (lo + hi) / 2
	return s.Query(2*node, lo, mid, ql, qr) + s.Query(2*node+1, mid+1, hi, ql, qr)
}
```

**Complexity:** update O(log N), query O(log N), space O(4N). The `4*n` sizing is the standard safe bound for the recursive layout.

---

### 4. LRU Cache (hashmap + doubly linked list)

One-sentence: O(1) `get`/`put` with least-recently-used eviction — a hashmap for lookup plus a doubly linked list for ordering; **this is one of the most-asked interview problems**, so memorize it.

**When you see…** "design an LRU cache", "evict least recently used", "O(1) get and put", "fixed-capacity cache" → **map + doubly linked list with sentinel head/tail**.

```go
// LRU Cache: hashmap (O(1) lookup) + doubly linked list (O(1) reorder).
// Sentinel head/tail remove all nil-edge handling. MRU sits next to head;
// the eviction victim sits next to tail.
type node struct {
	key, val   int
	prev, next *node
}

type LRUCache struct {
	cap        int
	m          map[int]*node
	head, tail *node // sentinels
}

func Constructor(capacity int) LRUCache {
	h, t := &node{}, &node{}
	h.next, t.prev = t, h
	return LRUCache{cap: capacity, m: make(map[int]*node), head: h, tail: t}
}

func (c *LRUCache) unlink(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (c *LRUCache) pushFront(n *node) { // insert right after head (most-recent)
	n.prev, n.next = c.head, c.head.next
	c.head.next.prev = n
	c.head.next = n
}

func (c *LRUCache) Get(key int) int {
	n, ok := c.m[key]
	if !ok {
		return -1
	}
	c.unlink(n)
	c.pushFront(n)
	return n.val
}

func (c *LRUCache) Put(key, value int) {
	if n, ok := c.m[key]; ok {
		n.val = value
		c.unlink(n)
		c.pushFront(n)
		return
	}
	if len(c.m) == c.cap {
		victim := c.tail.prev // least-recently-used
		c.unlink(victim)
		delete(c.m, victim.key)
	}
	n := &node{key: key, val: value}
	c.m[key] = n
	c.pushFront(n)
}
```

> The victim node **must store its own `key`** so you can `delete` it from the map on eviction. Sentinels mean no `if n == nil` checks anywhere. `container/list` works too, but a hand-rolled DLL is the point of the exercise.

**Complexity:** get O(1), put O(1), space O(capacity).

---

### 5. Sweep Line / Interval Scheduling (Meeting Rooms II)

One-sentence: to find peak concurrency / minimum resources over intervals, **sweep events in time order** — either two separately-sorted start/end arrays, or a min-heap of end times.

**When you see…** "minimum number of meeting rooms", "max overlapping intervals", "can attend all meetings", "CPU/platform/room scheduling", "merge/insert intervals" → **sort by start, then sweep (two arrays or a heap)**.

```go
import (
	"container/heap"
	"sort"
)

// Technique A: two sorted arrays (starts, ends) swept with two pointers.
func minRoomsTwoArrays(intervals [][]int) int {
	n := len(intervals)
	starts, ends := make([]int, n), make([]int, n)
	for i, iv := range intervals {
		starts[i], ends[i] = iv[0], iv[1]
	}
	sort.Ints(starts)
	sort.Ints(ends)
	rooms, maxRooms, j := 0, 0, 0
	for i := 0; i < n; i++ {
		if starts[i] < ends[j] { // starts before earliest end → need a room
			rooms++
			maxRooms = max(maxRooms, rooms)
		} else { // a meeting ended; reuse its room
			j++
		}
	}
	return maxRooms
}

// Technique B: min-heap of end times. Pop any room freed before this start;
// heap size at the end = peak concurrent meetings.
type minHeap []int

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *minHeap) Pop() any          { o := *h; n := len(o); x := o[n-1]; *h = o[:n-1]; return x }

func minRoomsHeap(intervals [][]int) int {
	sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
	h := &minHeap{}
	for _, iv := range intervals {
		if h.Len() > 0 && (*h)[0] <= iv[0] { // earliest-ending room is free
			heap.Pop(h)
		}
		heap.Push(h, iv[1])
	}
	return h.Len()
}
```

> Both are O(N log N) (the sort dominates). The heap form generalizes cleanly when you must also track *which* resource is reused (store an id alongside the end time).

**Complexity:** O(N log N) time, O(N) space.

---

### 6. Min-Stack + a note on ordered structures

One-sentence: a stack that also returns its current minimum in O(1) by keeping a **parallel "mins" stack** that records the running minimum at each level.

**When you see…** "stack that supports getMin in O(1)", "max in a stack", "design min stack" → **auxiliary min/max stack**.

```go
// MinStack: O(1) push/pop/top/getMin via a parallel "mins" stack.
type MinStack struct {
	data []int
	mins []int
}

func NewMinStack() *MinStack { return &MinStack{} }

func (s *MinStack) Push(x int) {
	s.data = append(s.data, x)
	if len(s.mins) == 0 || x < s.mins[len(s.mins)-1] {
		s.mins = append(s.mins, x)
	} else {
		s.mins = append(s.mins, s.mins[len(s.mins)-1])
	}
}

func (s *MinStack) Pop() {
	s.data = s.data[:len(s.data)-1]
	s.mins = s.mins[:len(s.mins)-1]
}

func (s *MinStack) Top() int    { return s.data[len(s.data)-1] }
func (s *MinStack) GetMin() int { return s.mins[len(s.mins)-1] }
```

**Complexity:** all ops O(1), space O(N).

#### Ordered structures in Go — there is no built-in balanced BST

Go's stdlib has **no `TreeMap`/`TreeSet`**: `map` is unordered and `container/list` is just a linked list. When an interview wants ordered operations, pick by what you actually need:

| Need | Idiomatic Go answer |
|------|---------------------|
| Repeated min/max extraction | `container/heap` (binary heap) |
| Ordered iteration, occasional inserts | **sorted slice** + `sort.Search` (binary search); insert with `slices.Insert` |
| Membership in sorted data | `sort.SearchInts` / `slices.BinarySearch` |
| Frequent ordered insert *and* delete | mention a 3rd-party tree (e.g. red-black/B-tree pkg) — note no stdlib option |

> Interview tip: say out loud "Go has no balanced BST in the stdlib, so I'd use a sorted slice with binary search for O(log N) lookup and O(N) insert, or a heap if I only need the extremes." That signals you know the language's limits.

---

## 14. String Algorithms

> All Go below is verified on Go 1.25 (`gofmt`/`go vet`/`go build` clean, runtime-asserted). Builtin `min`/`max` (Go 1.21+) used throughout.

### Quick reference

| Algorithm | Use | Time | Space |
|---|---|---|---|
| KMP | Single-pattern search, no hashing | O(N+M) | O(M) |
| Rabin-Karp | Search, multi-pattern, rolling hash | O(N+M) avg, O(NM) worst | O(1) |
| Z-algorithm | Prefix-match lengths, pattern search | O(N+M) | O(N+M) |
| Manacher | Longest palindromic substring | O(N) | O(N) |

---

### KMP (Knuth-Morris-Pratt)

What/when: linear substring search that never re-examines matched chars, using a prefix/failure function (LPS = longest proper prefix that is also suffix).

**When you see…** "find pattern in text", "substring index", "string matching without hashing", "repeated prefix", "failure function", "shortest repeating unit / period of string".

```go
// buildLPS: lps[i] = length of longest proper prefix of p[:i+1] that is also a suffix.
func buildLPS(p string) []int {
	lps := make([]int, len(p))
	k := 0
	for i := 1; i < len(p); i++ {
		for k > 0 && p[i] != p[k] {
			k = lps[k-1]
		}
		if p[i] == p[k] {
			k++
		}
		lps[i] = k
	}
	return lps
}

// kmpSearch: first index of p in s, or -1.
func kmpSearch(s, p string) int {
	if len(p) == 0 {
		return 0
	}
	lps := buildLPS(p)
	k := 0
	for i := 0; i < len(s); i++ {
		for k > 0 && s[i] != p[k] {
			k = lps[k-1]
		}
		if s[i] == p[k] {
			k++
		}
		if k == len(p) {
			return i - len(p) + 1 // continue with k = lps[k-1] to find all
		}
	}
	return -1
}
```

- Complexity: **O(N+M)** time, **O(M)** space.
- Period trick: if `n % (n-lps[n-1]) == 0`, the string is built from a repeating block of length `n-lps[n-1]`.

---

### Rabin-Karp (rolling hash)

What/when: hash the pattern and each window of text; slide the window in O(1) by removing the leading char and adding the trailing char. Always verify on hash match (collisions).

**When you see…** "rolling hash", "multiple patterns at once", "find duplicate substrings", "longest duplicate substring", "anagram windows by hash", "polynomial hash mod prime".

```go
func rabinKarp(s, p string) int {
	n, m := len(s), len(p)
	if m == 0 {
		return 0
	}
	if m > n {
		return -1
	}
	const base, mod = 256, 1_000_000_007
	var hp, hs, pow int64 = 0, 0, 1
	for i := 0; i < m; i++ {
		hp = (hp*base + int64(p[i])) % mod
		hs = (hs*base + int64(s[i])) % mod
		if i < m-1 {
			pow = (pow * base) % mod // base^(m-1)
		}
	}
	for i := 0; i+m <= n; i++ {
		if hp == hs && s[i:i+m] == p { // verify: hash equal != strings equal
			return i
		}
		if i+m < n {
			hs = ((hs-int64(s[i])*pow)%mod*base + int64(s[i+m])) % mod
			hs = (hs%mod + mod) % mod // keep non-negative after subtraction
		}
	}
	return -1
}
```

- Complexity: **O(N+M)** average, **O(N·M)** adversarial worst (all hashes collide). Space **O(1)**.
- Use a large prime `mod` and pick `base ≥ alphabet size`. Go subtraction can go negative under `%`, so re-normalize with `(x%mod+mod)%mod`.

---

### Z-algorithm (Z-array)

What/when: `z[i]` = length of the longest substring starting at `i` that matches a prefix of `s`. Build once in O(N); use for pattern search via `p + sep + s`.

**When you see…** "match against prefix", "Z-array", "pattern matching linear", "count occurrences", "string periodicity".

```go
func zArray(s string) []int {
	n := len(s)
	z := make([]int, n)
	if n == 0 {
		return z
	}
	z[0] = n
	l, r := 0, 0 // current [l,r] Z-box (rightmost match window)
	for i := 1; i < n; i++ {
		if i < r {
			z[i] = min(r-i, z[i-l])
		}
		for i+z[i] < n && s[z[i]] == s[i+z[i]] {
			z[i]++
		}
		if i+z[i] > r {
			l, r = i, i+z[i]
		}
	}
	return z
}

// zSearch: first index of p in s using a separator not in either string.
func zSearch(s, p string) int {
	concat := p + "\x00" + s
	z := zArray(concat)
	for i := len(p) + 1; i < len(concat); i++ {
		if z[i] >= len(p) {
			return i - len(p) - 1
		}
	}
	return -1
}
```

- Complexity: **O(N+M)** time, **O(N+M)** space.

---

### Manacher (longest palindromic substring)

What/when: find the longest palindromic substring in linear time by transforming the string with `#` separators (so even/odd cases unify) and reusing a mirror radius around the current center.

**When you see…** "longest palindromic substring", "all palindromic substrings count", "palindrome in O(n)".

```go
func longestPalindrome(s string) string {
	if len(s) == 0 {
		return ""
	}
	// Transform "aba" -> "^#a#b#a#$"; sentinels ^ and $ avoid bounds checks.
	t := make([]byte, 0, 2*len(s)+3)
	t = append(t, '^')
	for i := 0; i < len(s); i++ {
		t = append(t, '#', s[i])
	}
	t = append(t, '#', '$')

	n := len(t)
	p := make([]int, n) // p[i] = palindrome radius in t centered at i
	center, right := 0, 0
	for i := 1; i < n-1; i++ {
		if i < right {
			p[i] = min(right-i, p[2*center-i]) // mirror
		}
		for t[i+p[i]+1] == t[i-p[i]-1] { // expand (sentinels stop it)
			p[i]++
		}
		if i+p[i] > right {
			center, right = i, i+p[i]
		}
	}
	maxLen, centerIdx := 0, 0
	for i := 1; i < n-1; i++ {
		if p[i] > maxLen {
			maxLen, centerIdx = p[i], i
		}
	}
	start := (centerIdx - maxLen) / 2 // map back to original index
	return s[start : start+maxLen]
}
```

- Complexity: **O(N)** time, **O(N)** space.
- `p[i]` in `t` equals the palindrome length in the original `s`. Verified: `longestPalindrome("babad")` -> `"bab"` (or `"aba"`), `"cbbd"` -> `"bb"`, `"forgeeksskeegfor"` -> `"geeksskeeg"`.

---

### Practical patterns (show up far more often)

These beat the fancy algorithms in real interviews. Most lowercase-ASCII problems reduce to a `[26]int` count array.

**Group anagrams** — sorted string OR 26-count as map key.

```go
func groupAnagrams(strs []string) [][]string {
	m := make(map[string][]string)
	for _, s := range strs {
		b := []byte(s)
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] }) // O(k log k) key
		m[string(b)] = append(m[string(b)], s)
	}
	res := make([][]string, 0, len(m))
	for _, g := range m {
		res = append(res, g)
	}
	return res
}
```

> Faster key: build a 26-count and stringify it (`O(k)` per word, no sort). Use it when k is large.

**Is anagram** — single `[26]int`, increment for a, decrement for b.

```go
func isAnagram(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var cnt [26]int
	for i := 0; i < len(a); i++ {
		cnt[a[i]-'a']++
		cnt[b[i]-'a']--
	}
	for _, c := range cnt {
		if c != 0 {
			return false
		}
	}
	return true
}
```

**Char frequency** — the workhorse for windows/anagrams/permutations.

```go
var cnt [26]int
for i := 0; i < len(s); i++ {
	cnt[s[i]-'a']++
}
```

**Palindrome two-pointer** — converge from both ends.

```go
func isPalindrome(s string) bool {
	i, j := 0, len(s)-1
	for i < j {
		if s[i] != s[j] {
			return false
		}
		i++
		j--
	}
	return true
}
```

> For "valid palindrome" variants: lowercase + skip non-alphanumeric inside the loop. With Unicode use `[]rune(s)` first.

**String-to-int (atoi) edge cases** — skip spaces, optional sign, overflow-clamp BEFORE multiplying.

```go
func myAtoi(s string) int {
	i, n := 0, len(s)
	for i < n && s[i] == ' ' { // 1. leading spaces
		i++
	}
	sign := 1
	if i < n && (s[i] == '+' || s[i] == '-') { // 2. optional sign
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	const intMax, intMin = 1<<31 - 1, -(1 << 31)
	res := 0
	for i < n && s[i] >= '0' && s[i] <= '9' { // 3. digits; stop at non-digit
		d := int(s[i] - '0')
		if res > (intMax-d)/10 { // 4. clamp before overflow
			if sign == 1 {
				return intMax
			}
			return intMin
		}
		res = res*10 + d
		i++
	}
	return sign * res
}
```

Edge cases to recite: leading/trailing spaces, `+`/`-`, no digits at all (`""` -> 0), overflow clamp to `INT_MAX`/`INT_MIN`, stop at first non-digit (`"42abc"` -> 42).

---

### Go string realities (don't get burned)

| Idiom | Meaning |
|---|---|
| `s[i]` | a **byte** (`uint8`), NOT a character |
| `len(s)` | number of **bytes**, not runes |
| `for i, r := range s` | `r` is a **rune** (decoded UTF-8 code point), `i` is byte offset |
| `[]rune(s)` | full decode to runes — needed for Unicode indexing/reversal |
| `[]byte(s)` | byte slice — fine for ASCII, mutable scratch |
| `[26]int{}` | count array for lowercase ASCII via `s[i]-'a'` |
| `strings.Builder` | O(1) amortized concat; avoid `+=` in loops (O(n²)) |

```go
// Reverse a string safely for Unicode:
r := []rune(s)
for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
	r[i], r[j] = r[j], r[i]
}
result := string(r)

// Build, don't concatenate:
var b strings.Builder
for _, w := range words {
	b.WriteString(w)
}
out := b.String()
```

- `s[i]-'a'` only works for known lowercase ASCII. For mixed/Unicode use a `map[rune]int`.
- Strings are immutable; convert to `[]byte`/`[]rune` to mutate, then back with `string(...)`.

---

## 15. Advanced Graph Algorithms

All snippets compile under Go 1.21+ (builtin `min`/`max`). The `DSU` type below is defined in the Union-Find section.

| Algorithm | Use for | Complexity |
|-----------|---------|------------|
| Kruskal / Prim (MST) | Cheapest set of edges connecting all nodes | `E log E` / `E log V` |
| Bellman-Ford | Shortest path with negative edges; detect neg cycle | `V·E` |
| Floyd-Warshall | All-pairs shortest path, dense / small `V` | `V³` |
| Bipartite (2-color) | Partition into two conflict-free sets | `V+E` |
| 0-1 BFS | Shortest path, edge weights only 0 or 1 | `V+E` |
| Kosaraju / Tarjan (SCC) | Maximal mutually-reachable groups in digraph | `V+E` |

### Minimum Spanning Tree

One sentence: a min-total-weight edge subset connecting every vertex with no cycle.

**When you see…** "connect all points/cities at minimum cost", "min cost to make graph connected", "remove max-weight edges keeping connectivity".

Kruskal — sort edges, greedily add if it joins two components (Union-Find). Best for sparse / edge-list input.

```go
type Edge struct{ u, v, w int }

func kruskal(n int, edges []Edge) (int, []Edge) {
	sort.Slice(edges, func(i, j int) bool { return edges[i].w < edges[j].w })
	dsu := NewDSU(n) // DSU defined in the Union-Find section
	total, mst := 0, make([]Edge, 0, n-1)
	for _, e := range edges {
		if dsu.Union(e.u, e.v) { // joins two different components
			total += e.w
			mst = append(mst, e)
			if len(mst) == n-1 {
				break
			}
		}
	}
	return total, mst
}
```

Prim — grow a tree from one node, always pulling the cheapest crossing edge via a min-heap. Best for dense / adjacency input.

```go
type primItem struct{ node, cost int }

type primHeap []primItem

func (h primHeap) Len() int            { return len(h) }
func (h primHeap) Less(i, j int) bool  { return h[i].cost < h[j].cost }
func (h primHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *primHeap) Push(x interface{}) { *h = append(*h, x.(primItem)) }
func (h *primHeap) Pop() interface{} {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}

// adj[u] = list of {neighbor, weight}
func prim(n int, adj [][]primItem) int {
	visited := make([]bool, n)
	h := &primHeap{{node: 0, cost: 0}}
	total := 0
	for h.Len() > 0 {
		it := heap.Pop(h).(primItem)
		if visited[it.node] {
			continue
		}
		visited[it.node] = true
		total += it.cost
		for _, nb := range adj[it.node] {
			if !visited[nb.node] {
				heap.Push(h, primItem{node: nb.node, cost: nb.cost})
			}
		}
	}
	return total
}
```

### Bellman-Ford — `O(V·E)`

One sentence: single-source shortest paths that tolerates negative edges and detects negative cycles.

**When you see…** "negative weights", "currency arbitrage", "detect if costs can loop infinitely lower", or Dijkstra is wrong because edges go negative.

Relax all edges `V-1` times; one more successful relaxation ⇒ a negative cycle exists.

```go
func bellmanFord(n int, edges []Edge, src int) ([]int, bool) {
	dist := make([]int, n)
	for i := range dist {
		dist[i] = math.MaxInt
	}
	dist[src] = 0
	for i := 0; i < n-1; i++ { // relax V-1 times
		for _, e := range edges {
			if dist[e.u] != math.MaxInt && dist[e.u]+e.w < dist[e.v] {
				dist[e.v] = dist[e.u] + e.w
			}
		}
	}
	for _, e := range edges { // extra pass: any relaxation => negative cycle
		if dist[e.u] != math.MaxInt && dist[e.u]+e.w < dist[e.v] {
			return dist, true
		}
	}
	return dist, false
}
```

### Floyd-Warshall — `O(V³)`

One sentence: all-pairs shortest paths via DP over "allowed intermediate vertices".

**When you see…** "shortest path between every pair", small dense graph (`V ≤ ~400`), transitive closure, or repeated queries.

`dist` starts as the adjacency matrix (0 on diagonal, `MaxInt` for no edge). The `k` loop must be outermost.

```go
func floydWarshall(dist [][]int) [][]int {
	n := len(dist)
	for k := 0; k < n; k++ {
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if dist[i][k] != math.MaxInt && dist[k][j] != math.MaxInt {
					dist[i][j] = min(dist[i][j], dist[i][k]+dist[k][j])
				}
			}
		}
	}
	return dist
}
```

### Bipartite Check / 2-Coloring — `O(V+E)`

One sentence: try to color the graph with two colors so no edge joins same-colored nodes.

**When you see…** "split into two groups", "is graph bipartite", "possible to divide so no two enemies/dislikes are together", odd-cycle detection.

BFS each component; conflict (neighbor already same color) ⇒ not bipartite.

```go
func isBipartite(adj [][]int) bool {
	n := len(adj)
	color := make([]int, n) // 0=uncolored, 1/-1 = colors
	for s := 0; s < n; s++ {
		if color[s] != 0 {
			continue
		}
		color[s] = 1
		queue := []int{s}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			for _, v := range adj[u] {
				if color[v] == 0 {
					color[v] = -color[u]
					queue = append(queue, v)
				} else if color[v] == color[u] {
					return false // same color on both ends => not bipartite
				}
			}
		}
	}
	return true
}
```

### 0-1 BFS — `O(V+E)`

One sentence: Dijkstra replaced by a deque when every edge weight is 0 or 1 — push 0-edges to the front, 1-edges to the back.

**When you see…** weights restricted to {0,1}, "minimum number of walls/flips to cross a grid", "free vs cost-1 moves".

```go
type edge01 struct{ to, w int } // w is 0 or 1

func zeroOneBFS(n int, adj [][]edge01, src int) []int {
	const INF = 1 << 30
	dist := make([]int, n)
	for i := range dist {
		dist[i] = INF
	}
	dist[src] = 0
	deque := []int{src}
	for len(deque) > 0 {
		u := deque[0]
		deque = deque[1:]
		for _, e := range adj[u] {
			if dist[u]+e.w < dist[e.to] {
				dist[e.to] = dist[u] + e.w
				if e.w == 0 {
					deque = append([]int{e.to}, deque...) // push front
				} else {
					deque = append(deque, e.to) // push back
				}
			}
		}
	}
	return dist
}
```

### Strongly Connected Components — `O(V+E)`

One sentence: partition a directed graph into maximal groups where every node reaches every other.

**When you see…** "mutually reachable", "cycles in a directed graph collapsed into super-nodes", condensation graph, 2-SAT.

- **Kosaraju** — two DFS passes (intuitive): order by finish time, then DFS the transpose in reverse order.
- **Tarjan** — single DFS using `low-link` values and a stack (one pass, slightly trickier but faster constant).

Kosaraju:

```go
func kosaraju(n int, adj [][]int) [][]int {
	visited := make([]bool, n)
	order := make([]int, 0, n)

	// 1st pass: push nodes by finish time
	var dfs1 func(u int)
	dfs1 = func(u int) {
		visited[u] = true
		for _, v := range adj[u] {
			if !visited[v] {
				dfs1(v)
			}
		}
		order = append(order, u)
	}
	for u := 0; u < n; u++ {
		if !visited[u] {
			dfs1(u)
		}
	}

	// build transpose graph
	radj := make([][]int, n)
	for u := 0; u < n; u++ {
		for _, v := range adj[u] {
			radj[v] = append(radj[v], u)
		}
	}

	// 2nd pass: DFS transpose in reverse finish order
	for i := range visited {
		visited[i] = false
	}
	var comp []int
	var dfs2 func(u int)
	dfs2 = func(u int) {
		visited[u] = true
		comp = append(comp, u)
		for _, v := range radj[u] {
			if !visited[v] {
				dfs2(v)
			}
		}
	}
	sccs := make([][]int, 0)
	for i := len(order) - 1; i >= 0; i-- {
		u := order[i]
		if !visited[u] {
			comp = nil
			dfs2(u)
			sccs = append(sccs, comp)
		}
	}
	return sccs
}
```

---

## 16. Advanced DP Patterns

| Pattern | State | Transition essence | Complexity |
|---------|-------|--------------------|------------|
| Interval DP | `dp[i][j]` over range `[i,j]` | split on a "last/center" point `k` inside | `N³` |
| Bitmask DP | `dp[mask]` / `dp[mask][i]` | toggle one set bit per step (`N ≤ ~20`) | `2ᴺ·N` |
| Tree DP | `dp[node]` (tuple via DFS) | combine children results | `N` |
| Digit DP | `dp[pos][state][tight]` | choose next digit `0..limit` | `digits·states·10` |

### Interval DP — `O(N³)`

State: `dp[i][j]` = best answer for sub-range `(i, j)`. Transition: pick a center/last element `k` between `i` and `j` and combine the two sides. Base: empty range = 0. Iterate by increasing range length.

**When you see…** "burst balloons", "matrix chain multiplication", "merge stones / min cost to combine adjacent", "remove boxes", anything where the order of resolving a range matters.

Example — Burst Balloons (`dp[i][j]` excludes endpoints; `k` is the *last* balloon burst):

```go
// dp[i][j] = max coins from bursting balloons strictly between i and j.
func maxCoins(nums []int) int {
	balloons := make([]int, len(nums)+2)
	balloons[0], balloons[len(balloons)-1] = 1, 1
	copy(balloons[1:], nums)
	n := len(balloons)
	dp := make([][]int, n)
	for i := range dp {
		dp[i] = make([]int, n)
	}
	for length := 2; length < n; length++ { // gap between i and j
		for i := 0; i+length < n; i++ {
			j := i + length
			for k := i + 1; k < j; k++ { // k = last balloon burst in (i,j)
				gain := balloons[i]*balloons[k]*balloons[j] + dp[i][k] + dp[k][j]
				dp[i][j] = max(dp[i][j], gain)
			}
		}
	}
	return dp[0][n-1]
}
```

### Bitmask DP — `O(2ᴺ·N)`, `N ≤ ~20`

State: `dp[mask]` where each bit = an element used/visited; often paired with a position `dp[mask][i]` for TSP (last node = `i`). Transition: extend `mask` by setting one unused bit. Base: `dp[0] = 1` (or 0 for cost).

**When you see…** small `N` (`≤ 20`), "assign N items to N slots", "visit all nodes once" (TSP / shortest Hamiltonian path), "subset partitioning", "minimum compatible groups".

Example — count perfect task↔person assignments (next person = popcount of mask):

```go
// dp[mask] = ways to fill people 0..(popcount(mask)-1) using task set mask.
func assignWays(compatible [][]bool) int {
	n := len(compatible)
	dp := make([]int, 1<<n)
	dp[0] = 1
	for mask := 0; mask < (1 << n); mask++ {
		person := popcount(mask) // next person to assign
		if person >= n {
			continue
		}
		for task := 0; task < n; task++ {
			if mask&(1<<task) == 0 && compatible[person][task] {
				dp[mask|(1<<task)] += dp[mask]
			}
		}
	}
	return dp[(1<<n)-1]
}

func popcount(x int) int { // or bits.OnesCount(uint(x))
	c := 0
	for x > 0 {
		x &= x - 1
		c++
	}
	return c
}
```

For TSP: `dp[mask][i]` = min cost path covering `mask`, ending at `i`; answer = `min over i of dp[full][i] (+ return edge)`.

### DP on Trees — `O(N)`

State: each DFS call returns a small tuple describing the subtree (e.g. `{include root, exclude root}`). Transition: fold children into the parent's tuple. Base: leaf returns `{val, 0}`.

**When you see…** "house robber III", "max independent set / min vertex cover on a tree", "max path sum in a tree", "tree diameter", any "choose nodes with parent/child constraint".

Example — House Robber III (max independent set on tree):

```go
type robPair struct{ with, without int } // rob this node vs skip it

func robTree(n int, children [][]int, val []int) int {
	var dfs func(u int) robPair
	dfs = func(u int) robPair {
		with, without := val[u], 0
		for _, c := range children[u] {
			cp := dfs(c)
			with += cp.without                  // if we rob u, skip children
			without += max(cp.with, cp.without) // else take best of child
		}
		return robPair{with, without}
	}
	r := dfs(0)
	return max(r.with, r.without)
}
```

### Digit DP — brief

Idea: count numbers in `[0, N]` satisfying a digit property by building the number digit-by-digit left→right. State: `pos` (index), problem-specific carry (e.g. `prev` digit), a `tight` flag (still bounded by `N`'s prefix), and `started` (skip leading zeros). Memoize only the non-tight states (tight states are unique per prefix).

**When you see…** "count numbers ≤ N with <some digit condition>", "sum of digits", "no repeated/adjacent digits", "contains digit d", huge ranges where brute force is impossible.

Skeleton — count integers in `[0, N]` with no two equal adjacent digits:

```go
func countNoAdjEqual(N int) int {
	digits := []int{}
	for N > 0 {
		digits = append([]int{N % 10}, digits...)
		N /= 10
	}
	if len(digits) == 0 {
		return 1 // just 0
	}
	memo := map[[3]int]int{}
	var dp func(pos, prev int, tight, started bool) int
	dp = func(pos, prev int, tight, started bool) int {
		if pos == len(digits) {
			return 1
		}
		key := [3]int{pos*100 + prev, b2i(tight), b2i(started)}
		if v, ok := memo[key]; ok && !tight {
			return v
		}
		limit := 9
		if tight {
			limit = digits[pos]
		}
		total := 0
		for d := 0; d <= limit; d++ {
			if started && d == prev {
				continue // adjacent equal digits not allowed
			}
			nStarted := started || d > 0
			nPrev := prev
			if nStarted {
				nPrev = d
			}
			total += dp(pos+1, nPrev, tight && d == limit, nStarted)
		}
		if !tight { // only memoize unbounded states
			memo[key] = total
		}
		return total
	}
	return dp(0, -1, true, false)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
```

### DP Optimization Awareness

| Technique | What it buys | When to reach for it |
|-----------|--------------|----------------------|
| Top-down (memo) | Only computes reachable states; easy to write from recurrence | Sparse state space, complex transitions, tree/digit DP |
| Bottom-up (tabulation) | No recursion overhead, predictable order, easy rolling | Dense states, when you can order dimensions cleanly |
| Rolling array | Cut space `O(N·M) → O(M)` when `dp[i]` depends only on `dp[i-1]` | 1D/knapsack/grid rows; iterate knapsack capacity descending for 0/1 |
| Prefix sums / monotonic deque | Drop a factor of `N` from range/window transitions | `dp[i] = min/sum over a window of dp[j]` |
| Encode state compactly | Fit state into an int/bitmask key | Bitmask DP, digit DP memo keys |

Rules of thumb: prefer **top-down** when the recurrence is obvious but the iteration order isn't (or many states are unreachable); prefer **bottom-up** when you also want space optimization via rolling arrays. Always define **state, transition, base** before coding.

---

## 17. Problem Drill & Worked Examples

> **How to use.** Part A is a recognition deck — scan the pattern tables, cover the approach column, and try to recall the one-liner from the problem name alone. Part B models out-loud reasoning for four escalating problems; rehearse the *verbalization*, not the code. Complexities use N = input size unless noted.

### Part A — Must-know problems by pattern

#### Arrays & Hashing

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Two Sum | Hashmap value→index, look up complement `target-x` while scanning | O(N) / O(N) |
| Group Anagrams | Key each word by sorted chars (or 26-count tuple), bucket in map | O(N·K log K) / O(N·K) |
| Top K Frequent Elements | Count map, then bucket sort by frequency (or size-K heap) | O(N) / O(N) |
| Product of Array Except Self | Prefix products L→R, then suffix pass R→L; no division | O(N) / O(1)* |
| Valid Sudoku | 9 sets each for rows/cols/boxes; box key `r/3,c/3` | O(1) / O(1) |
| Longest Consecutive Sequence | Put all in set; start a run only at numbers with no `x-1` | O(N) / O(N) |
| Contains Duplicate | Set; return true on first re-seen element | O(N) / O(N) |
| Encode/Decode Strings | Length-prefix each chunk: `len#str` | O(N) / O(N) |

\*output array excluded from extra space.

#### Two Pointers

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Valid Palindrome | Skip non-alphanumerics, compare lowercased ends inward | O(N) / O(1) |
| Two Sum II (sorted) | L/R pointers, move by sum vs target | O(N) / O(1) |
| 3Sum | Sort; fix `i`, two-pointer the rest, skip dups | O(N²) / O(1) |
| Container With Most Water | L/R ends; area = min height × width, move shorter side | O(N) / O(1) |
| Trapping Rain Water | Two pointers tracking `leftMax`/`rightMax`, add deficit on shorter side | O(N) / O(1) |

#### Sliding Window

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Best Time to Buy/Sell Stock | Track min price so far, max (price − min) | O(N) / O(1) |
| Longest Substring Without Repeating | Window + last-seen map; jump left past duplicate | O(N) / O(min(N,Σ)) |
| Longest Repeating Char Replacement | Window valid while `len − maxCount ≤ k`; shrink else | O(N) / O(1) |
| Permutation in String | Fixed-size window, compare 26-count arrays vs target | O(N) / O(1) |
| Minimum Window Substring | Expand to satisfy need-map, then shrink while valid | O(N) / O(Σ) |
| Sliding Window Maximum | Monotonic decreasing deque of indices | O(N) / O(K) |

#### Stack

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Valid Parentheses | Push opens, match/pop on closes, empty at end | O(N) / O(N) |
| Min Stack | Pair each value with running min (or second min-stack) | O(1) ops / O(N) |
| Evaluate Reverse Polish Notation | Push operands, pop two on operator | O(N) / O(N) |
| Daily Temperatures | Monotonic decreasing stack of indices, resolve on warmer day | O(N) / O(N) |
| Car Fleet | Sort by position desc; stack of arrival times, merge if caught | O(N log N) / O(N) |
| Largest Rectangle in Histogram | Monotonic increasing stack; on pop, width = gap to prev | O(N) / O(N) |
| Generate Parentheses | Backtrack with open/close counts (`open<n`, `close<open`) | O(4ⁿ/√n) / O(N) |

#### Binary Search

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Binary Search | Classic lo/hi midpoint narrowing | O(log N) / O(1) |
| Search in Rotated Sorted Array | Find sorted half each step, decide which side holds target | O(log N) / O(1) |
| Find Minimum in Rotated Sorted Array | Compare `mid` to `hi` to pick unsorted half | O(log N) / O(1) |
| Koko Eating Bananas | Binary search on answer (eat-rate); feasibility check | O(N log M) / O(1) |
| Search a 2D Matrix | Treat as flattened sorted array, one binary search | O(log MN) / O(1) |
| Time-Based Key-Value Store | Per-key sorted list of (ts,val); binary search ≤ ts | O(log N) / O(N) |
| Median of Two Sorted Arrays | Binary search partition of smaller array | O(log min(M,N)) / O(1) |

#### Linked List

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Reverse Linked List | Iterate flipping `next` with prev/cur | O(N) / O(1) |
| Merge Two Sorted Lists | Dummy head, splice smaller node each step | O(N+M) / O(1) |
| Reorder List | Find mid (fast/slow), reverse 2nd half, weave | O(N) / O(1) |
| Remove Nth Node From End | Two pointers `n` apart, delete when lead hits end | O(N) / O(1) |
| Linked List Cycle | Floyd fast/slow; meet ⇒ cycle | O(N) / O(1) |
| Copy List With Random Pointer | Interleave clones (or map orig→copy), then split | O(N) / O(1) |
| Add Two Numbers | Walk both with carry, build digit list | O(N) / O(1) |
| Merge K Sorted Lists | Min-heap of heads (or pairwise merge) | O(N log K) / O(K) |
| LRU Cache | Hashmap + doubly linked list; move-to-front on access | O(1) ops / O(C) |

#### Trees

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Invert Binary Tree | Swap children recursively | O(N) / O(H) |
| Maximum Depth | `1 + max(left,right)` | O(N) / O(H) |
| Diameter of Binary Tree | Post-order height; update best = `lH+rH` | O(N) / O(H) |
| Balanced Binary Tree | Height check returning −1 sentinel on imbalance | O(N) / O(H) |
| Same Tree | Compare nodes in lockstep | O(N) / O(H) |
| Subtree of Another Tree | At each node test sameTree against subroot | O(N·M) / O(H) |
| Lowest Common Ancestor (BST) | Walk down by value vs both targets | O(H) / O(1) |
| Binary Tree Level Order | BFS queue, one level per outer loop | O(N) / O(N) |
| Right Side View | BFS, take last node per level | O(N) / O(N) |
| Validate BST | DFS carrying (low, high) bounds | O(N) / O(H) |
| Kth Smallest in BST | In-order traversal, stop at k-th | O(H+k) / O(H) |
| Construct from Preorder & Inorder | Preorder gives root; split inorder by its index | O(N) / O(N) |
| Serialize/Deserialize | Pre-order with null markers; rebuild from queue | O(N) / O(N) |
| Binary Tree Max Path Sum | Post-order; gain = `max(0, child)`, best across split | O(N) / O(H) |

#### Tries

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Implement Trie | Node with children map/array + `isEnd` flag | O(L) ops / O(Σ·N·L) |
| Add and Search Word (`.` wildcard) | DFS over children when char is `.` | O(L) / O(N·L) |
| Word Search II | Build trie of words, DFS the grid pruning by trie | O(M·4·3ᴸ) / O(N·L) |

#### Heap / Priority Queue

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Kth Largest Element in Array | Size-K min-heap (or quickselect) | O(N log K) / O(K) |
| Kth Largest in Stream | Maintain size-K min-heap; top is answer | O(log K)/add / O(K) |
| Top K Frequent Elements | Count map → heap of size K | O(N log K) / O(N) |
| Find Median From Data Stream | Max-heap (low half) + min-heap (high half), balanced | O(log N)/add / O(N) |
| Task Scheduler | Greedy by max freq; idle = `(maxFreq−1)·(n+1)+ties` | O(N) / O(1) |
| Merge K Sorted Lists | Min-heap of list heads | O(N log K) / O(K) |
| K Closest Points to Origin | Max-heap of size K by distance² | O(N log K) / O(K) |

#### Backtracking

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Subsets | At each index choose include/exclude | O(N·2ᴺ) / O(N) |
| Combination Sum (reuse) | DFS, stay on index to allow repeats, prune on overshoot | O(2^T) / O(T) |
| Combinations | Pick from `start..n`, recurse `start+1` | O(K·C(N,K)) / O(K) |
| Permutations | Swap-in-place or used[] array | O(N·N!) / O(N) |
| Word Search | DFS grid 4-dir, mark visited, backtrack | O(M·4·3ᴸ) / O(L) |
| Palindrome Partitioning | DFS cut points; recurse if prefix is palindrome | O(N·2ᴺ) / O(N) |
| Subsets II / Combination Sum II | Sort, skip duplicate siblings (`i>start && a[i]==a[i-1]`) | O(N·2ᴺ) / O(N) |
| N-Queens | Place row-by-row; track col + two diagonals sets | O(N!) / O(N) |

#### Graphs

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Number of Islands | DFS/BFS flood-fill each unvisited land cell | O(MN) / O(MN) |
| Clone Graph | DFS/BFS with old→new node map | O(V+E) / O(V) |
| Pacific Atlantic Water Flow | BFS inward from each ocean border, intersect reachable | O(MN) / O(MN) |
| Course Schedule | Topological sort / cycle detection (Kahn or DFS colors) | O(V+E) / O(V+E) |
| Redundant Connection | Union-Find; first edge that unions same set | O(N·α) / O(N) |
| Word Ladder | BFS over words, generate `*`-pattern neighbors | O(N·L²) / O(N·L) |
| Rotting Oranges | Multi-source BFS from rotten cells, count minutes | O(MN) / O(MN) |
| Network Delay Time | Dijkstra from source, answer = max finalized dist | O(E log V) / O(V+E) |
| Walls and Gates | Multi-source BFS from all gates | O(MN) / O(MN) |
| Graph Valid Tree | N−1 edges and one connected component (Union-Find) | O(N·α) / O(N) |

#### Dynamic Programming — 1D

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Climbing Stairs | `dp[i]=dp[i-1]+dp[i-2]` (Fibonacci) | O(N) / O(1) |
| House Robber | `dp[i]=max(dp[i-1], dp[i-2]+a[i])` | O(N) / O(1) |
| House Robber II | Run linear robber twice excluding first / last house | O(N) / O(1) |
| Coin Change | `dp[a]=min(dp[a−c]+1)` over coins (unbounded) | O(A·C) / O(A) |
| Longest Increasing Subsequence | Patience: binary-search tails array | O(N log N) / O(N) |
| Word Break | `dp[i]` true if some `dp[j]` true and `s[j:i]` in dict | O(N²) / O(N) |
| Decode Ways | `dp[i]` from 1-digit and valid 2-digit prefixes | O(N) / O(1) |
| Maximum Product Subarray | Track running max **and** min (negatives flip) | O(N) / O(1) |
| Partition Equal Subset Sum | Subset-sum to `total/2`, boolean DP | O(N·S) / O(S) |

#### Dynamic Programming — 2D

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Unique Paths | `dp[i][j]=dp[i-1][j]+dp[i][j-1]` | O(MN) / O(N) |
| Longest Common Subsequence | Match ⇒ `1+diag`; else `max(up,left)` | O(MN) / O(N) |
| Edit Distance | min of insert/delete/replace; +0 if chars match | O(MN) / O(N) |
| 0/1 Knapsack | `dp[w]=max(dp[w], dp[w−wt]+val)`, iterate weight desc | O(N·W) / O(W) |
| Longest Palindromic Substring | Expand around each center (2N−1 centers) | O(N²) / O(1) |
| Palindromic Substrings (count) | Same expand-around-center, count each | O(N²) / O(1) |
| Coin Change II (count ways) | Unbounded knapsack, coins as outer loop | O(A·C) / O(A) |
| Regular Expression Matching | `dp[i][j]`; `*` = zero or more of prev char | O(MN) / O(MN) |
| Interleaving String | `dp[i][j]` reachable from up or left if char matches | O(MN) / O(N) |

#### Greedy & Intervals

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Merge Intervals | Sort by start; merge if `cur.start ≤ last.end` | O(N log N) / O(N) |
| Insert Interval | Add lefts, merge overlaps with new, add rights | O(N) / O(N) |
| Non-Overlapping Intervals | Sort by end; greedily keep earliest-ending, count removals | O(N log N) / O(1) |
| Meeting Rooms | Sort by start; any overlap ⇒ false | O(N log N) / O(1) |
| Meeting Rooms II | Sort starts & ends; sweep counting concurrent | O(N log N) / O(N) |
| Jump Game | Track farthest reach; fail if `i > reach` | O(N) / O(1) |
| Jump Game II | BFS-style: extend current jump boundary, count hops | O(N) / O(1) |
| Gas Station | If total gas ≥ cost, restart start when tank goes negative | O(N) / O(1) |
| Hand of Straights | Count map; greedily form runs from smallest card | O(N log N) / O(N) |

#### Bit Manipulation

| Problem | Approach (one line) | Time / Space |
|---------|--------------------|:------------:|
| Single Number | XOR all; pairs cancel | O(N) / O(1) |
| Number of 1 Bits | `n &= n-1` clears lowest set bit, count loops | O(1) / O(1) |
| Counting Bits | `dp[i] = dp[i>>1] + (i&1)` | O(N) / O(N) |
| Reverse Bits | Shift result left, OR in lowest bit of input, 32× | O(1) / O(1) |
| Missing Number | XOR indices with values (or sum formula) | O(N) / O(1) |
| Sum of Two Integers | `a^b` sum-without-carry, `(a&b)<<1` carry, loop | O(1) / O(1) |

---

### Part B — Worked examples: how to think

Each walkthrough is the script you say out loud: **clarify → brute force (+ cost) → key insight → optimal → complexity → edge cases**.

#### B1 (Easy / Hashmap) — Two Sum

1. **Clarify.** Exactly one solution? Same element reused? Sorted? Return indices or values? → assume one answer, no reuse, unsorted, return indices.
2. **Brute force.** Two nested loops checking every pair for `a[i]+a[j]==target`. **O(N²) time, O(1) space.**
3. **Key insight.** For each `x` I only need to know whether `target−x` was already seen — that's an O(1) membership + lookup question.
4. **Optimal.** One pass: Go `map[int]int` of value→index. For each `x`, check if `target−x` is in the map; if yes return both indices, else store `x`.
5. **Complexity.** O(N) time, O(N) space.
6. **Edge cases.** Duplicate values (`[3,3]`, target 6) — works because we check *before* inserting; no pair exists — loop ends with no return.

#### B2 (Medium / Sliding Window) — Longest Substring Without Repeating Characters

1. **Clarify.** Charset (ASCII vs unicode)? Substring (contiguous) not subsequence? Return length or the string? → ASCII, contiguous, return length.
2. **Brute force.** Check every substring for uniqueness with a set. **O(N²) (or O(N³) naïvely) time, O(min(N,Σ)) space.**
3. **Key insight.** A valid window can only grow on the right; when a repeat enters, the left edge never needs to move backward — just jump it past the duplicate's last position.
4. **Optimal.** Right pointer scans; `map[byte]int` holds last-seen index. On a repeat inside the window, set `left = max(left, lastSeen+1)`. Track `max(right−left+1)`.
5. **Complexity.** O(N) time (each pointer advances ≤ N), O(min(N, Σ)) space.
6. **Edge cases.** Empty string → 0; all identical chars (`"bbbb"`) → 1; repeat *before* current left (`"abba"`) — the `max` guard stops left from regressing.

#### B3 (Medium / BFS-Graph) — Rotting Oranges

1. **Clarify.** Grid values 0 empty / 1 fresh / 2 rotten? Rot spreads to 4-neighbors per minute? Return −1 if any fresh remains? → yes to all.
2. **Brute force.** Repeatedly scan the whole grid each minute, rotting neighbors, until no change. **O((MN)²) time** worst case.
3. **Key insight.** All currently-rotten cells spread *simultaneously* each minute — that's exactly multi-source BFS, where BFS depth = elapsed minutes.
4. **Optimal.** Seed a queue with all rotten cells and count fresh ones. Process level by level (size-bounded loop), rotting fresh neighbors, incrementing minutes per non-empty level. At end, fresh==0 ? minutes : −1.
5. **Complexity.** O(MN) time, O(MN) space for the queue.
6. **Edge cases.** No fresh oranges at start → 0 minutes; a fresh orange unreachable (walled by empties) → return −1; grid with only empty cells → 0.

#### B4 (Hard-ish / DP) — Coin Change (fewest coins to make amount)

1. **Clarify.** Unlimited coins per denomination? Return *count* (not combinations)? Impossible ⇒ −1? → unbounded, count, −1 on failure.
2. **Brute force.** Recurse subtracting each coin from `amount`, try all paths. **Exponential O(C^amount)** with heavy overlap.
3. **Key insight.** `minCoins(amount)` depends only on `minCoins(amount−coin)` — overlapping subproblems ⇒ DP over amounts (unbounded knapsack).
4. **Optimal.** `dp` slice size `amount+1`, init to `amount+1` (sentinel "infinity"), `dp[0]=0`. For each `a` from 1..amount, `dp[a]=min(dp[a], dp[a−coin]+1)` over coins where `coin ≤ a`. Answer `dp[amount]` unless still sentinel ⇒ −1.
5. **Complexity.** O(amount × #coins) time, O(amount) space.
6. **Edge cases.** `amount==0` → 0; no coin divides amount (`coins=[2]`, amount 3) → −1; single coin equal to amount → 1.

> **Why bottom-up here:** interviewers like the iterative table because it removes recursion-depth risk and makes the O(amount·C) bound obvious; mention memoized top-down as the equivalent if you reason there first.

---

## 18. Interview-Day Gotchas & Tips

### Process (say these out loud — communication is graded)
1. **Restate** the problem and confirm with 1 example.
2. **Clarify**: input size? sorted? negatives/duplicates/empty? in-place? return value vs print?
3. **Brute force first** — state its complexity, *then* optimize. A working O(N²) beats a broken O(N).
4. **Name the pattern** ("this looks like sliding window because…").
5. **Dry-run** your code on the example before claiming done.
6. **State complexity** (time AND space) unprompted.

### Edge cases to always probe (your CLAUDE.md's "8 edge cases")
- empty input / single element
- all identical / all distinct
- negatives, zero, overflow boundaries
- already sorted / reverse sorted
- duplicates
- target not present / not found
- max-size input (does it fit time budget?)
- special chars / Unicode (for strings — Go strings are UTF-8 bytes!)

### Go-specific landmines
- **`range` over a string** yields **runes** (Unicode code points) with byte offsets, not bytes. `s[i]` yields a **byte**. For ASCII problems use `s[i]` (byte); for Unicode use `[]rune(s)`.
- **Slices share backing arrays** — copy before mutating if the original must survive (Section 2).
- **Map iteration order is random** — sort keys for deterministic output.
- **Integer division truncates toward zero**; `-7 / 2 == -3`, and `%` can be negative: `-7 % 3 == -1`. For always-positive modulo: `((x % m) + m) % m`.
- **No built-in `abs` for ints**; Go 1.21+ has built-in `min`/`max` for ordered types.
- **Recursion has no TCO** — deep recursion (≥ ~10⁵ frames) can blow the stack; convert to iterative with an explicit stack if depth is huge.
- **`container/heap` needs you to implement `Push`/`Pop` on a pointer receiver**; forget the pointer and it won't mutate.
- **nil slice vs empty slice**: both have len 0; `append` works on a `nil` slice. Returning `nil` is idiomatic for "no results".

### Time-management
- ~5 min clarify + brute force, ~10 min optimal approach + complexity, ~20 min code, ~10 min test/dry-run. If stuck >5 min, **say your approach aloud** — interviewers nudge.
- If the optimal eludes you, **code the brute force cleanly** and articulate the optimization you *would* make. Partial + clear communication often passes.

### The "I'm blanking" recovery checklist
1. What's the brute force? Code that.
2. What's the bottleneck operation? (repeated search → hash/heap; repeated min/max range → segment tree/monotonic; recomputation → memoize.)
3. Is the input sorted or can I sort it? (unlocks two pointers / binary search.)
4. Can I trade space for time? (hash map of seen values.)
5. Scan the Section 9 table against the keywords in the prompt.

---

*Built for fast revision. If a section feels thin when you revisit, that's the cue to drill 2–3 problems on that pattern and thicken your template here.*
