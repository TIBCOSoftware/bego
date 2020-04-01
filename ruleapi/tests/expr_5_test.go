package tests

import (
	"context"
	"testing"

	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/ruleapi"

	"github.com/stretchr/testify/assert"
)

//1 arithmetic operation
func Test_5_Expr(t *testing.T) {

	actionCount := map[string]int{"count": 0}
	rs, err := createRuleSession(t)
	assert.Nil(t, err)
	r1 := ruleapi.NewRule("r1")
	err = r1.AddExprCondition("c1", "(($.t1.p1 + $.t2.p1) == 5) && (($.t1.p2 > $.t2.p2) && ($.t1.p3 == $.t2.p3))", nil)
	assert.Nil(t, err)
	r1.SetActionService(createActionServiceFromFunction(t, a5))
	r1.SetContext(actionCount)

	err = rs.AddRule(r1)
	assert.Nil(t, err)

	err = rs.Start(nil)
	assert.Nil(t, err)

	var ctx context.Context

	t1, err := model.NewTupleWithKeyValues("t1", "t1")
	assert.Nil(t, err)
	err = t1.SetInt(nil, "p1", 1)
	assert.Nil(t, err)
	err = t1.SetDouble(nil, "p2", 1.3)
	assert.Nil(t, err)
	err = t1.SetString(nil, "p3", "t3")
	assert.Nil(t, err)

	ctx = context.WithValue(context.TODO(), TestKey{}, t)
	err = rs.Assert(ctx, t1)
	assert.Nil(t, err)

	t2, err := model.NewTupleWithKeyValues("t2", "t2")
	assert.Nil(t, err)
	err = t2.SetInt(nil, "p1", 4)
	assert.Nil(t, err)
	err = t2.SetDouble(nil, "p2", 1.1)
	assert.Nil(t, err)
	err = t2.SetString(nil, "p3", "t3")
	assert.Nil(t, err)

	ctx = context.WithValue(context.TODO(), TestKey{}, t)
	err = rs.Assert(ctx, t2)
	assert.Nil(t, err)
	deleteRuleSession(t, rs, t1)
	count := actionCount["count"]
	if count != 1 {
		t.Errorf("expected [%d], got [%d]\n", 1, count)
	}
}

func a5(ctx context.Context, rs model.RuleSession, ruleName string, tuples map[model.TupleType]model.Tuple, ruleCtx model.RuleContext) {
	t := ctx.Value(TestKey{}).(*testing.T)
	t.Logf("Test_5_Expr executed!")
	actionCount := ruleCtx.(map[string]int)
	count := actionCount["count"]
	actionCount["count"] = count + 1
}
