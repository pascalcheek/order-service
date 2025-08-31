package cache

import (
	"sync"

	"order-service/internal/model"
)

type cacheNode struct {
	order model.Order
	next  *cacheNode
	prev  *cacheNode
}

type Cache struct {
	mu       sync.RWMutex
	orders   map[string]*cacheNode
	capacity int
	count    int
	head     *cacheNode
	tail     *cacheNode
}

func New(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 1000 // default capacity
	}
	return &Cache{
		orders:   make(map[string]*cacheNode),
		capacity: capacity,
	}
}

func (c *Cache) Set(order model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.orders[order.OrderUID]; exists {
		node.order = order
		c.moveToFront(node)
		return
	}

	node := &cacheNode{
		order: order,
	}

	if c.count >= c.capacity {
		c.removeOldest()
	}

	c.orders[order.OrderUID] = node
	c.addToFront(node)
	c.count++
}

func (c *Cache) Get(orderUID string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	node, exists := c.orders[orderUID]
	if !exists {
		return model.Order{}, false
	}

	go c.updateAccess(node)

	return node.order, true
}

func (c *Cache) updateAccess(node *cacheNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[node.order.OrderUID]; exists {
		c.moveToFront(node)
	}
}

func (c *Cache) GetAll() []model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]model.Order, 0, len(c.orders))
	for _, node := range c.orders {
		orders = append(orders, node.order)
	}
	return orders
}

func (c *Cache) LoadFromOrders(orders []model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders = make(map[string]*cacheNode)
	c.head = nil
	c.tail = nil
	c.count = 0

	for i := 0; i < len(orders) && i < c.capacity; i++ {
		order := orders[i]
		node := &cacheNode{
			order: order,
		}
		c.orders[order.OrderUID] = node
		c.addToFront(node)
		c.count++
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

func (c *Cache) Capacity() int {
	return c.capacity
}

func (c *Cache) addToFront(node *cacheNode) {
	if c.head == nil {
		c.head = node
		c.tail = node
		return
	}

	node.next = c.head
	c.head.prev = node
	c.head = node
}

func (c *Cache) moveToFront(node *cacheNode) {
	if node == c.head {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if node == c.tail {
		c.tail = node.prev
	}

	node.next = c.head
	node.prev = nil
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

func (c *Cache) removeOldest() {
	if c.tail == nil {
		return
	}

	delete(c.orders, c.tail.order.OrderUID)

	if c.tail.prev != nil {
		c.tail.prev.next = nil
	} else {
		c.head = nil
	}
	c.tail = c.tail.prev
	c.count--
}
