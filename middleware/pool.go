package middleware

import (
	"reflect"
	"errors"
	"fmt"
	"sync"
)

//实体池
type Pool interface {
	Take() (Entity, error)
	//归还实体
	Return(pl Entity) error
	Total() uint32
	Used() uint32
}

type Entity interface {
	Id() uint32
}

type myPool struct {
	total       uint32
	etype       reflect.Type
	getEntity   func() Entity
	container   chan Entity
	idContainer map[uint32]bool
	mux         sync.Mutex
}

func (this *myPool) Take() (Entity, error) {
	entity, ok := <-this.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	this.idContainer[entity.Id()] = false
	return entity, nil
}

func (this *myPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}
	if this.etype != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is NOT %s\n", this.etype)
		return errors.New(errMsg)
	}
	entityId := entity.Id()
	canResult := this.compareAndSetForIdContainer(entityId, false, true)
	if canResult == 1 {
		this.container <- entity
		return nil
	} else if canResult == 0 {
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entityId)
		return errors.New(errMsg)
	} else {
		errMsg := fmt.Sprintf("The entity (id=%d) is illegal!\n", entityId)
		return errors.New(errMsg)
	}

}

func (this *myPool) Total() uint32 {
	return this.total
}

func (this *myPool) Used() uint32 {
	return this.total - uint32(len(this.container))
}

//比较设置实体ID容器中与给定实体ID对应的键值对的元素值
// -1键值对不存在
// 0 操作失败
// 1操作成功
func (this *myPool) compareAndSetForIdContainer(entityId uint32, oldValue bool, newValue bool) int8 {
	this.mux.Lock()
	defer this.mux.Unlock()
	v, ok := this.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	this.idContainer[entityId] = newValue
	return 1
}

func NewPool(total uint32,
	entitytype reflect.Type,
	getEntity func() Entity) (Pool, error) {
	if total == 0 {
		errMsg := fmt.Sprintf("The pool can not be initialized! (tatol=%d)\n", total)
		return nil, errors.New(errMsg)
	}

	size := int(total)
	container := make(chan Entity, size)
	idContainer := make(map[uint32]bool)
	for i := 0; i < size; i++ {
		newEntity := getEntity()
		if entitytype != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("The type of result of func getEntity() is NOT %s\n", entitytype)
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}
	pool := &myPool{
		total:       total,
		etype:       entitytype,
		getEntity:   getEntity,
		container:   container,
		idContainer: idContainer,
	}
	return pool, nil
}
