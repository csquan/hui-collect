package utils

import (
	"fmt"
	snowflake2 "github.com/bwmarrin/snowflake"
	"sync"
	"time"
)

type Node struct {
	node *snowflake2.Node
}

var node *snowflake2.Node
var once sync.Once

// NewIDGenerator 生成一个新的 Node 类型变量
func NewIDGenerator() *Node {
	once.Do(func() {
		node, _ = snowflake2.NewNode(0)
	})
	return &Node{
		node: node,
	}
}

func (n *Node) Generate() string {
	now := time.Now().Format("20060102")
	return fmt.Sprintf("%v%v", now, n.node.Generate())
}
