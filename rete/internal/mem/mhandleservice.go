package mem

import (
	"sync"

	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/rete/internal/types"
)

type handleServiceImpl struct {
	types.NwServiceImpl
	allHandles map[string]types.ReteHandle
	sync.RWMutex
}

func NewHandleCollection(nw types.Network, config map[string]interface{}) types.HandleService {
	hc := handleServiceImpl{}
	hc.Nw = nw
	hc.allHandles = make(map[string]types.ReteHandle)
	return &hc
}

func (hc *handleServiceImpl) Init() {
}

func (hc *handleServiceImpl) RemoveHandle(tuple model.Tuple) types.ReteHandle {
	hc.Lock()
	defer hc.Unlock()
	rh, found := hc.allHandles[tuple.GetKey().String()]
	if found {
		delete(hc.allHandles, tuple.GetKey().String())
		return rh
	}
	return nil
}

func (hc *handleServiceImpl) GetHandle(tuple model.Tuple) types.ReteHandle {
	hc.RLock()
	defer hc.RUnlock()
	return hc.allHandles[tuple.GetKey().String()]
}

func (hc *handleServiceImpl) GetHandleByKey(key model.TupleKey) types.ReteHandle {
	hc.RLock()
	defer hc.RUnlock()
	return hc.allHandles[key.String()]
}

func (hc *handleServiceImpl) GetOrCreateHandle(nw types.Network, tuple model.Tuple) types.ReteHandle {
	hc.Lock()
	defer hc.Unlock()
	h, found := hc.allHandles[tuple.GetKey().String()]
	if !found {
		h = newReteHandleImpl(nw, tuple)

		hc.allHandles[tuple.GetKey().String()] = h //[tuple.GetKey().String()] = h
	}
	return h
}
