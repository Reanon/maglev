package maglevhash

import "errors"

// MaglevHash refer to https://www.jianshu.com/p/9a9b269e68f7
type MaglevHash struct {
	n           int // size of node
	m           int // size of the lookup table
	permutation [][]int
	entry       []int  // lookup table
	nodeState   []bool // node list
}

func NewMaglevHash(n int, m int) *MaglevHash {
	// m := findPrime(5 * n)
	permutation := make([][]int, n)
	entry := make([]int, m)
	nodeState := make([]bool, n)
	for idx, _ := range nodeState {
		nodeState[idx] = true
	}

	return &MaglevHash{
		m:           m,
		n:           n,
		permutation: permutation,
		entry:       entry,
		nodeState:   nodeState,
	}
}

// Permutate 生成 n * m 大小的优先填充表
func (mh *MaglevHash) Permutate() {
	for idx, _ := range mh.permutation {
		mh.permutation[idx] = make([]int, mh.m)
	}
	for i := 0; i < mh.n; i++ {
		offset := Hash1(i+1) % mh.m
		skip := Hash2(i+1)%(mh.m-1) + 1
		for j := 0; j < mh.m; j++ {
			mh.permutation[i][j] = (offset + j*skip) % mh.m
		}
	}
}

// Populate 填充查找表
func (mh *MaglevHash) Populate() {
	// 查找表初始化
	for idx, _ := range mh.entry {
		mh.entry[idx] = -1
	}
	// 记录某个节点填充表的下一个值
	next := make([]int, mh.n)
	n := 0
	for {
		for i := 0; i < mh.n; i++ {
			// 如果节点失效，就跳过
			if !mh.nodeState[i] {
				continue
			}
			c := mh.permutation[i][next[i]]
			for mh.entry[c] >= 0 {
				next[i]++
				c = mh.permutation[i][next[i]]
			}
			mh.entry[c] = i
			next[i]++
			n++
			// 填充完毕就跳出
			if n == mh.m {
				return
			}
		}
	}
}

// DownNode 节点下线
func (mh *MaglevHash) DownNode(idx int) error {
	if idx > mh.n-1 {
		return errors.New("invalid idx")
	}
	mh.nodeState[idx] = false
	return nil
}

// UpNode 节点上线
func (mh *MaglevHash) UpNode(idx int) error {
	if idx > mh.n-1 {
		return errors.New("invalid idx")
	}
	mh.nodeState[idx] = true
	return nil
}
