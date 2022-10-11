package tests

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/project-flogo/rules/common"
	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/config"
	"github.com/project-flogo/rules/ruleapi"
	"github.com/stretchr/testify/assert"
)

func createRuleSession() (model.RuleSession, error) {
	rs, _ := ruleapi.GetOrCreateRuleSession("test")

	tupleDescFileAbsPath := common.GetAbsPathForResource("src/github.com/project-flogo/rules/ruleapi/tests/tests.json")

	dat, err := ioutil.ReadFile(tupleDescFileAbsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = model.RegisterTupleDescriptors(string(dat))
	if err != nil {
		return nil, err
	}
	return rs, nil
}

//conditions and actions
func trueCondition(ruleName string, condName string, tuples map[model.TupleType]model.Tuple, ctx model.RuleContext) bool {
	return true
}
func falseCondition(ruleName string, condName string, tuples map[model.TupleType]model.Tuple, ctx model.RuleContext) bool {
	return false
}
func emptyAction(ctx context.Context, rs model.RuleSession, ruleName string, tuples map[model.TupleType]model.Tuple, ruleCtx model.RuleContext) {

}

func printTuples(t *testing.T, oprn string, tupleMap map[string]map[string]model.Tuple) {

	for k, v := range tupleMap {
		t.Logf("%s tuples for type [%s]\n", oprn, k)
		for k1 := range v {
			t.Logf("    tuples key [%s]\n", k1)
		}
	}
}
func printModified(t *testing.T, modified map[string]map[string]model.RtcModified) {

	for k, v := range modified {
		t.Logf("%s tuples for type [%s]\n", "Modified", k)
		for k1 := range v {
			t.Logf("    tuples key [%s]\n", k1)
		}
	}
}

type txnCtx struct {
	Testing *testing.T
	TxnCnt  int
}

func createActionServiceFromFunction(t *testing.T, actionFunction model.ActionFunction) model.ActionService {
	fname := runtime.FuncForPC(reflect.ValueOf(actionFunction).Pointer()).Name()
	cfg := &config.ServiceDescriptor{
		Name:        fname,
		Description: fname,
		Type:        config.TypeServiceFunction,
		Function:    actionFunction,
	}
	aService, err := ruleapi.NewActionService(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, aService)
	return aService
}

type TestKey struct{}

func Drain(port string) {
	for {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("", port), time.Second)
		if conn != nil {
			conn.Close()
		}
		if err != nil && strings.Contains(err.Error(), "connect: connection refused") {
			break
		}
	}
}

func Pour(port string) {
	for {
		conn, _ := net.Dial("tcp", net.JoinHostPort("", port))
		if conn != nil {
			conn.Close()
			break
		}
	}
}

func CaptureStdOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func Command(name string, arg ...string) {
	_, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
}
