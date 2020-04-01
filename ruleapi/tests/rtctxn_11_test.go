package tests

import (
	"context"
	"testing"

	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/ruleapi"

	"github.com/stretchr/testify/assert"
)

//1 rtc->one assert triggers two rule actions each rule action asserts 2 more tuples.Verify Tuple type and Tuples count.
func Test_T11(t *testing.T) {

	rs, err := createRuleSession(t)
	assert.Nil(t, err)

	rule := ruleapi.NewRule("R11")
	err = rule.AddCondition("R11_c1", []string{"t1.none"}, trueCondition, nil)
	assert.Nil(t, err)
	rule.SetActionService(createActionServiceFromFunction(t, r11_action))
	rule.SetPriority(1)
	err = rs.AddRule(rule)
	assert.Nil(t, err)
	t.Logf("Rule added: [%s]\n", rule.GetName())

	rule1 := ruleapi.NewRule("R112")
	err = rule1.AddCondition("R112_c1", []string{"t1.none"}, trueCondition, nil)
	assert.Nil(t, err)
	rule1.SetActionService(createActionServiceFromFunction(t, r112_action))
	rule1.SetPriority(1)
	err = rs.AddRule(rule1)
	assert.Nil(t, err)
	t.Logf("Rule added: [%s]\n", rule1.GetName())

	txnCtx := txnCtx{t, 0}
	rs.RegisterRtcTransactionHandler(t11Handler, &txnCtx)
	err = rs.Start(nil)
	assert.Nil(t, err)

	t1, err := model.NewTupleWithKeyValues("t1", "t10")
	assert.Nil(t, err)
	err = rs.Assert(context.TODO(), t1)
	assert.Nil(t, err)

	t2, err := model.NewTupleWithKeyValues("t3", "t2")
	assert.Nil(t, err)
	t3, err := model.NewTupleWithKeyValues("t3", "t1")
	assert.Nil(t, err)

	deleteRuleSession(t, rs, t1, t2, t3)

}

func r11_action(ctx context.Context, rs model.RuleSession, ruleName string, tuples map[model.TupleType]model.Tuple, ruleCtx model.RuleContext) {
	t1 := tuples[model.TupleType("t1")].(model.MutableTuple)
	id, _ := t1.GetString("id")

	if id == "t10" {
		t2, _ := model.NewTupleWithKeyValues("t3", "t2")
		rs.Assert(ctx, t2)
	}
}

func r112_action(ctx context.Context, rs model.RuleSession, ruleName string, tuples map[model.TupleType]model.Tuple, ruleCtx model.RuleContext) {
	t1 := tuples[model.TupleType("t1")].(model.MutableTuple)
	id, _ := t1.GetString("id")

	if id == "t10" {
		t3, _ := model.NewTupleWithKeyValues("t3", "t1")
		rs.Assert(ctx, t3)
	}
}

func t11Handler(ctx context.Context, rs model.RuleSession, rtxn model.RtcTxn, handlerCtx interface{}) {
	if done {
		return
	}

	txnCtx := handlerCtx.(*txnCtx)
	txnCtx.TxnCnt = txnCtx.TxnCnt + 1
	t := txnCtx.Testing
	if txnCtx.TxnCnt == 1 {
		lA := len(rtxn.GetRtcAdded())
		if lA != 2 {
			t.Errorf("RtcAdded: Types expected [%d], got [%d]\n", 2, lA)
			printTuples(t, "Added", rtxn.GetRtcAdded())
		} else {
			tuples, _ := rtxn.GetRtcAdded()["t1"]
			if tuples != nil {
				if len(tuples) != 1 {
					t.Errorf("RtcAdded: Expected [%d], got [%d]\n", 1, len(tuples))
					printTuples(t, "Added", rtxn.GetRtcAdded())
				}
			}
			tuples3, _ := rtxn.GetRtcAdded()["t3"]
			if tuples3 != nil {
				if len(tuples3) != 2 {
					t.Errorf("RtcAdded: Expected [%d], got [%d]\n", 2, len(tuples3))
					printTuples(t, "Added", rtxn.GetRtcAdded())
				}
			}
		}
		lM := len(rtxn.GetRtcModified())
		if lM != 0 {
			t.Errorf("RtcModified: Expected [%d], got [%d]\n", 0, lM)
			printModified(t, rtxn.GetRtcModified())
		}
		lD := len(rtxn.GetRtcDeleted())
		if lD != 0 {
			t.Errorf("RtcDeleted: Expected [%d], got [%d]\n", 0, lD)
			printTuples(t, "Deleted", rtxn.GetRtcDeleted())
		}
	}
}
