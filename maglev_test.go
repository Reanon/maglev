package maglevhash

import (
	"log"
	"testing"
)

const sizeN = 4096
const lookupSizeM = 4099 // need prime

// TestRemoveConsistent 测试删除一个节点之后的均衡度
func TestRemoveConsistent(t *testing.T) {
	maglevHash := NewMaglevHash(sizeN, lookupSizeM)
	maglevHash.Permutate()
	maglevHash.Populate()
	table1 := make([]int, lookupSizeM, lookupSizeM)
	copy(table1, maglevHash.entry)
	log.Println("lookup:", table1)
	// 移除节点
	err := maglevHash.DownNode(4095)
	if err != nil {
		log.Println("down error", err)
		return
	}
	maglevHash.Permutate()
	maglevHash.Populate()
	table2 := maglevHash.entry
	log.Println("lookup:", table2)
	consistent := 0
	for i := 0; i < lookupSizeM; i++ {
		if table1[i] == table2[i] {
			consistent += 1
		}
	}
	log.Println("consistent rate:", float64(consistent)/lookupSizeM*100)
}

// TestAddConsistent 测试增加一个节点之后的均衡度
func TestAddConsistent(t *testing.T) {
	nodeIdx := 1118
	maglevHash := NewMaglevHash(sizeN, lookupSizeM)
	// 移除节点
	err := maglevHash.DownNode(nodeIdx)
	if err != nil {
		log.Println("down error", err)
		return
	}
	maglevHash.Permutate()
	maglevHash.Populate()
	table1 := make([]int, lookupSizeM, lookupSizeM)
	copy(table1, maglevHash.entry)
	log.Println("lookup:", table1)
	// 移除节点
	err = maglevHash.UpNode(nodeIdx)
	if err != nil {
		log.Println("down error", err)
		return
	}
	maglevHash.Permutate()
	maglevHash.Populate()
	table2 := maglevHash.entry
	log.Println("lookup:", table2)
	consistent := 0
	for i := 0; i < lookupSizeM; i++ {
		if table1[i] == table2[i] {
			consistent += 1
		}
	}
	log.Println("consistent rate:", float64(consistent)/lookupSizeM*100)
}
