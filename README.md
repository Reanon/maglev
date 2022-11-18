# Google Maglev Hashing实现

## 背景

Maglev 是 Google 开发的基于 kernal bypass 技术实现的 4 层负载均衡，它具有非常强大的负载性能，承载了Google绝大部分接入流量。Maglev在负载均衡算法上采用自行开发的一致性哈希算法被称为Maglev Hashing，该哈希算法在节点变化时能够尽量少的影响其他几点，且尽可能的保证负载的均衡，是一个非常优秀的一致性哈希算法。

下面想用 Golang 做一个 Maglev 简单的实现。

## 原理说明

Maglev Hashing 的基本思想是为每一个节点生成一个优先填充表，列表的值就是该节点想要填充查询表的位置.

<img src="https://aliyun-typora-img.oss-cn-beijing.aliyuncs.com/imgs/202211181131381.png" alt="img" style="zoom: 50%;" />

如 Table1 所示，节点B0，会按照顺序 3、0，4，...依次去尝试填充查询表。实际上，所有的节点会轮流按照各自优先列表的值填充查询表。也就是说，每个节点都有几乎均等的机会根据优先表来填充查询表，直到查询表被填满。

当出现节点变化，如B1宕机时，查询表会重新生成，因为节点的优先填充表不变，所以B0和B2原来的填充位置不变，B1宕机后确实的位置被B0和B2瓜分，按照轮流填充的机制，B0和B2基本也是均衡的。

## 算法实现

设 `M` 为查询表的大小。对与每一个节点 `i`，`permutation[i]`为优先填充表，`permutation[i] `的取值是数组`[0, M-1]`的一个随机顺序排列，`permutation`是一个二维数组。

### 辅助函数

下面介绍论文给出的高效生成`permutation[i]`的方法:

- 首先使用两种哈希函数来哈希节点生成两个数字，`offset` 和 ` skip`。
- 论文中是计算节点名称的哈希值，为了简单这里就直接计算了节点的索引值，

哈希函数用的是**算法导论**里提到的乘法散列法，代码如下：

- 第二个哈希函数只是修改了一个参数值，哈希算法是一样的。

```go
func Hash1(k int) int {
    s := uint64(2654435769)
    p := uint32(14)
    tmp := (s * uint64(k)) % (1 << 32)
    return int(tmp / (1 << (32 - p)))
}

func Hash2(k int) int {
    s := uint64(1654435769)
    p := uint32(14)
    tmp := (s * uint64(k)) % (1 << 32)
    return int(tmp / (1 << (32 - p)))
}
```

`offset` 和 ` skip`计算方式如下：

```
offset <- h1(name[i]) mod M
skip <- h2(name[i]) mod (M−1)+1
```

从而得到 permutation[i] 中每一个值的计算方式：

- 注意：M必须为质数，这样才能尽可能保证 skip 与 M 互斥。

```c
permutation[i][j] <- (offset+ j×skip) mod M , 0<= j <= M-1
```


这里要寻找合适的质数 M 我使用了简单的筛选算法：

```go
func isPrime(n int) bool {
    if n < 2 {
        return false
    }
    end := int(math.Sqrt(float64(n)))
    for i := 2; i <= end; i++ {
        if n%i == 0 {
            return false
        }
    }
    return true
}

func findPrime(n int) int {
    // 始终有大于n的质数
    for {
        if isPrime(n) {
            return n
        }
        n++
    }
}
```

### 具体实现

> 上面介绍了一些辅助函数，下面介绍算法的具体实现流程：

定义一个结构MaglevHash和结构体生成函数，golang的标准实现。其中permutation为一个 N*M 的二维数组，entry为长度N的查询表，nodeState为长度N的记录节点时候的下线的表。

```go
type MaglevHash struct {
    m, n        int
    permutation [][]int
    entry       []int
    nodeState   []bool
}

func NewMaglevHash(n int) *MaglevHash {
    m := findPrime(5 * n)
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
```



接下来是生成permutation的函数，计算节点时实际上传入的是节点索引值加一，避免传入0，影响哈希值的计算：

```go
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
```

生成好节点优先填充表之后，就可以根据该表填充查询表：

```go
func (mh *MaglevHash) Populate() {
    for idx, _ := range mh.entry {
        mh.entry[idx] = -1
    }
    next := make([]int, mh.n)
    n := 0
    for {
        for i := 0; i < mh.n; i++ {
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
            if n == mh.m {
                return
            }
        }
    }
}
```

在填充查询表时，会检查节点是否下线，若节点下线，则会忽略该节点。

```go
func (mh *MaglevHash) DownNode(idx int) error {
    if idx > mh.n-1 {
        return errors.New("invalid idx")
    }
    mh.nodeState[idx] = false
    return nil
}
```

节点下线时，需要调用该函数，然后再调用`Populate()`重新填充查询表。

至此，Maglev hashing 一个简单的实现就算完成了，后续希望使用生产环境的哈希函数来替换本文用到哈希函数，并考虑在nginx上实现该一致性哈希算法。



## 参考资料

1、[Google Maglev Hashing实现](https://www.jianshu.com/p/9a9b269e68f7)

2、[一致性哈希算法 | 春水煎茶](https://writings.sh/post/consistent-hashing-algorithms-part-1-the-problem-and-the-concept)