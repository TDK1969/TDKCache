package consistenthash

import (
	"fmt"
	"testing"
)

/*
410719615 (node3)
450215437 (2)
498629140 (6)
582255842 (node2)
1026140150 (node1)
1296257913 (node4)
1437840500 (node3)
1842515611 (3)
1870407145 (node2)
2212294583 (1)
2226203566 (5)
2250211548 (node4)
2753623628 (node2)
3149780312 (node1)
3419873751 (node4)
3542599386 (node3)
4088798008 (4)
4134892627 (node1)
*/

func TestConsistentHashGet(t *testing.T) {
	r := NewHashRing(nil, 3)
	r.Add("node1", "node2", "node3", "node4")
	testCases := map[string]string{
		"1": "node4",
		"2": "node2",
		"3": "node2",
		"4": "node1",
		"5": "node4",
		"6": "node2",
	}
	hashMap := make(map[string]uint32)
	for k, v := range testCases {
		hashMap[k] = r.hash([]byte(k))
		if rv := r.Get(k); rv != v {
			fmt.Printf("key [%s] should get from node [%s], but get [%s]\n", k, v, rv)
		}
	}
}

func TestConsistentHashAdd(t *testing.T) {
	r := NewHashRing(nil, 3)
	r.Add("node1", "node2", "node3")
	testCases := map[string]string{
		"1": "node2",
		"2": "node2",
		"3": "node2",
		"4": "node1",
		"5": "node2",
		"6": "node2",
	}
	for k, v := range testCases {
		if rv := r.Get(k); rv != v {
			fmt.Printf("before add, key [%s] should get from node [%s], but get [%s]\n", k, v, rv)
		}
	}
	r.Add("node4")
	testCases["1"] = "node4"
	testCases["5"] = "node4"
	for k, v := range testCases {
		if rv := r.Get(k); rv != v {
			fmt.Printf("after add , ey [%s] should get from node [%s], but get [%s]\n", k, v, rv)
		}
	}
}

func TestConsistentHashDel(t *testing.T) {
	r := NewHashRing(nil, 3)
	r.Add("node1", "node2", "node3", "node4")
	testCases := map[string]string{
		"1": "node4",
		"2": "node2",
		"3": "node2",
		"4": "node1",
		"5": "node4",
		"6": "node2",
	}
	for k, v := range testCases {
		if rv := r.Get(k); rv != v {
			fmt.Printf("before delete, key [%s] should get from node [%s], but get [%s]\n", k, v, rv)
		}
	}
	r.Del("node2")
	testCases["2"] = "node1"
	testCases["6"] = "node1"
	testCases["3"] = "node4"
	for k, v := range testCases {
		if rv := r.Get(k); rv != v {
			fmt.Printf("after delete , ey [%s] should get from node [%s], but get [%s]\n", k, v, rv)
		}
	}
}
