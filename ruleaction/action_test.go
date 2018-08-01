package ruleaction

import (
	"github.com/TIBCOSoftware/bego/ruleapi"
	"testing"
	"github.com/TIBCOSoftware/bego/common/model"
	"strconv"
	"io/ioutil"
	"log"
	"fmt"
	"time"
)


func TestAction (t *testing.T) {
	rs := ruleapi.GetOrCreateRuleSession("asession")

	dat, err := ioutil.ReadFile("/home/bala/go/src/github.com/TIBCOSoftware/bego/common/model/tupledescriptor.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("desc [%s]\n", string(dat))

	rs.RegisterTupleDescriptors(string(dat))

	//fmt.Printf(string(dat))
	loadRules(rs)



	for i := 1; i < 2  ; i++ {
		debit := model.NewStreamTuple(model.TupleTypeAlias("debitevent"))
		debit.SetString(nil, rs,"name", "Bob")
		fs := strconv.FormatFloat(float64(i*100), 'E', -1, 32)
		debit.SetString(nil, rs,"debit", fs)
		rs.Assert(nil, debit)
	}

	st1 := model.NewStreamTuple(model.TupleTypeAlias("customerevent"))
	st1.SetString (nil, rs,"name", "Bob")
	st1.SetString (nil, rs,"status", "active")
	st1.SetFloat (nil, rs,"balance", 1000)
	rs.Assert(nil, st1)
}

func TestActionTwo (t *testing.T) {
	rs := ruleapi.GetOrCreateRuleSession("asession")

	dat, err := ioutil.ReadFile("/home/bala/go/src/github.com/TIBCOSoftware/bego/common/model/tupledescriptor.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("desc [%s]\n", string(dat))

	rs.RegisterTupleDescriptors(string(dat))

	//fmt.Printf(string(dat))
	loadRules(rs)

	st1 := model.NewStreamTuple(model.TupleTypeAlias("customerevent"))
	st1.SetString (nil, rs,"name", "Bob")
	st1.SetString (nil, rs,"status", "active")
	st1.SetFloat  (nil, rs,"balance", 1000)
	rs.Assert(nil, st1)


	for i := 1; i < 2  ; i++ {
		debit := model.NewStreamTuple(model.TupleTypeAlias("debitevent"))
		debit.SetString(nil, rs,"name", "Bob")
		fs := strconv.FormatFloat(float64(i*100), 'E', -1, 32)
		debit.SetString(nil, rs,"debit", fs)
		rs.Assert(nil, debit)
	}

}

func TestActionThree (t *testing.T) {

	e := &EventTimer{151}

	scheduleTask(5, e)

	time.Sleep(time.Minute)
}

type Task interface {
	performOps()
}

type EventTimer struct {
	x int
}

func (e *EventTimer) performOps() {
	fmt.Printf ("The task is to print myself [%d]\n", e.x)
}

func scheduleTask (period int, t Task) *time.Timer {

	tmr := time.AfterFunc(time.Second * time.Duration(period), func() {
		t.performOps()
	})

	return tmr
}