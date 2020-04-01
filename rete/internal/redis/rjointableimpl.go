package redis

import (
	"context"
	"strconv"

	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/redisutils"
	"github.com/project-flogo/rules/rete/internal/types"
)

type joinTableImpl struct {
	types.NwElemIdImpl
	redisutils.RedisHdl
	idr   []model.TupleType
	rule  model.Rule
	name  string
	jtKey string
}

func newJoinTableImpl(nw types.Network, handle redisutils.RedisHdl, rule model.Rule, identifiers []model.TupleType, name string) types.JoinTable {
	jt := joinTableImpl{
		RedisHdl: handle,
	}
	jt.initJoinTableImpl(nw, rule, identifiers, name)
	return &jt
}

func (jt *joinTableImpl) initJoinTableImpl(nw types.Network, rule model.Rule, identifiers []model.TupleType, name string) {
	jt.SetID(nw)
	jt.idr = identifiers
	jt.rule = rule
	jt.name = name
	jt.jtKey = nw.GetPrefix() + ":" + "jt:" + name
}

func (jt *joinTableImpl) AddRow(handles []types.ReteHandle) types.JoinTableRow {
	row := newJoinTableRow(jt.RedisHdl, jt.jtKey, handles, jt.Nw)
	for i := 0; i < len(row.GetHandles()); i++ {
		handle := row.GetHandles()[i]
		jt.Nw.GetJtRefService().AddEntry(handle, jt.name, row.GetID())
	}
	row.Write()
	return row
}

func (jt *joinTableImpl) RemoveRow(rowID int) types.JoinTableRow {
	row := jt.GetRow(nil, rowID)
	rowId := strconv.Itoa(rowID)
	jt.HDel(jt.jtKey, rowId)
	return row
}

func (jt *joinTableImpl) RemoveAllRows(ctx context.Context) {
	rowIter := jt.GetRowIterator(ctx)
	for rowIter.HasNext() {
		row := rowIter.Next()
		//first, from jTable, remove row
		jt.RemoveRow(row.GetID())
		for _, hdl := range row.GetHandles() {
			jt.Nw.GetJtRefService().RemoveEntry(hdl, jt.GetName(), row.GetID())
		}
		//Delete the rowRef itself
		rowIter.Remove()
	}
}

func (jt *joinTableImpl) GetRowCount() int {
	return jt.HLen(jt.name)
}

func (jt *joinTableImpl) GetRule() model.Rule {
	return jt.rule
}

func (jt *joinTableImpl) GetRowIterator(ctx context.Context) types.JointableRowIterator {
	return newRowIterator(ctx, jt.RedisHdl, jt)
}

func (jt *joinTableImpl) GetRow(ctx context.Context, rowID int) types.JoinTableRow {
	key := jt.HGet(jt.jtKey, strconv.Itoa(rowID))
	rowId := strconv.Itoa(rowID)
	return createRow(ctx, jt.RedisHdl, jt.name, rowId, key.(string), jt.Nw)
}

func (jt *joinTableImpl) GetName() string {
	return jt.name
}

type rowIteratorImpl struct {
	ctx    context.Context
	iter   *redisutils.MapIterator
	jtName string
	nw     types.Network
	curr   types.JoinTableRow
	redisutils.RedisHdl
}

func newRowIterator(ctx context.Context, handle redisutils.RedisHdl, jTable types.JoinTable) types.JointableRowIterator {
	key := jTable.GetNw().GetPrefix() + ":jt:" + jTable.GetName()
	ri := rowIteratorImpl{}
	ri.ctx = ctx
	ri.iter = handle.GetMapIterator(key)
	ri.nw = jTable.GetNw()
	ri.jtName = jTable.GetName()
	ri.RedisHdl = handle
	return &ri
}

func (ri *rowIteratorImpl) HasNext() bool {
	return ri.iter.HasNext()
}

func (ri *rowIteratorImpl) Next() types.JoinTableRow {
	rowId, key := ri.iter.Next()
	tupleKeyStr := key.(string)
	ri.curr = createRow(ri.ctx, ri.RedisHdl, ri.jtName, rowId, tupleKeyStr, ri.nw)
	return ri.curr
}

func (ri *rowIteratorImpl) Remove() {
	ri.iter.Remove()
}
