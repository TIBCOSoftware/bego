package rulesapp

import (
	"context"
	"fmt"
	"testing"

	"github.com/tibmatt/bego/common/model"
	"github.com/tibmatt/bego/ruleapi"
)

func TestTwo(t *testing.T) {

	fmt.Println("** Welcome to BEGo **")

	//Create Rule, define conditiond and set action callback
	rule := ruleapi.NewRule("Name is Bob")
	fmt.Printf("Rule added: [%s]\n", rule.GetName())
	rule.AddCondition("c1", []model.TupleType{"n1"}, checkForBob) // check for name "Bob" in n1
	rule.SetAction(bobRuleFired)
	rule.SetPriority(1)

	//Create Rule, define conditiond and set action callback
	rule2 := ruleapi.NewRule("Bobs age is 35")
	fmt.Printf("Rule added: [%s]\n", rule2.GetName())
	rule2.AddCondition("c1", []model.TupleType{"n1"}, checkForBobAge) // check for name "Bob" in n1
	rule2.SetAction(bobAgeRuleFired)
	rule2.SetPriority(2)

	//Create a RuleSession and add the above Rule
	ruleSession := ruleapi.GetOrCreateRuleSession("testsession")
	ruleSession.AddRule(rule)
	ruleSession.AddRule(rule2)
	// ruleSession.AddRule(rule3)

	// ctx := context.Background()

	//Now assert a few facts and see if the Rule Action callback fires.
	fmt.Println("Asserting n1 tuple with name=Bob")
	tuple1 := model.NewTuple("n1")
	tuple1.SetString(nil, "name", "Bob")
	tuple1.SetInt(nil, "age", 35)

	ruleSession.Assert(nil, tuple1)

	//Retract them
	ruleSession.Retract(nil, tuple1)

	//You may delete the rule
	ruleSession.DeleteRule(rule.GetName())

}

func checkForBobAge(ruleName string, condName string, tuples map[model.TupleType]model.Tuple) bool {
	//This conditions filters on name="Bob"
	tuple := tuples["n1"]
	if tuple == nil {
		fmt.Println("Should not get a nil tuple in FilterCondition! This is an error")
		return false
	}
	age := tuple.GetInt("age")
	return age == 35
}

func bobAgeRuleFired(ctx context.Context, rs model.RuleSession, ruleName string, tuples map[model.TupleType]model.Tuple) {
	fmt.Printf("Rule [%s] fired\n", ruleName)
	tuple1 := tuples["n1"]
	if tuple1 == nil {
		fmt.Println("Should not get nil tuples here in Action! This is an error")
	}
	name1 := tuple1.GetString("name")
	fmt.Printf("n1.name = [%s]\n", name1)
	// assertTom(ctx, rs)
}
