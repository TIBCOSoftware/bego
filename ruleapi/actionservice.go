package ruleapi

import (
	"context"
	"fmt"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/rules/common/model"
	"github.com/project-flogo/rules/config"
)

var logger = log.ChildLogger(log.RootLogger(), "rules")

// rule action service
type ruleActionService struct {
	Name     string
	Type     string
	Function model.ActionFunction
	Act      activity.Activity
	Input    map[string]interface{}
}

// NewActionService creates new rule action service
func NewActionService(serviceCfg *config.ServiceDescriptor) (model.ActionService, error) {

	raService := &ruleActionService{
		Name:  serviceCfg.Name,
		Type:  serviceCfg.Type,
		Input: make(map[string]interface{}),
	}

	switch serviceCfg.Type {
	default:
		return nil, fmt.Errorf("service type - '%s' is not supported", serviceCfg.Type)
	case "":
		return nil, fmt.Errorf("service type can't be empty")
	case config.TypeServiceFunction:
		if serviceCfg.Function == nil {
			return nil, fmt.Errorf("service function can't empty")
		}
		raService.Function = serviceCfg.Function
	case config.TypeServiceActivity:
		// inflate activity from ref
		if serviceCfg.Ref[0] == '#' {
			var ok bool
			activityRef := serviceCfg.Ref
			serviceCfg.Ref, ok = support.GetAliasRef("activity", activityRef)
			if !ok {
				return nil, fmt.Errorf("activity '%s' not imported", activityRef)
			}
		}

		act := activity.Get(serviceCfg.Ref)
		if act == nil {
			return nil, fmt.Errorf("unsupported Activity:" + serviceCfg.Ref)
		}

		f := activity.GetFactory(serviceCfg.Ref)

		if f != nil {
			initCtx := newInitContext(serviceCfg.Name, serviceCfg.Settings, log.ChildLogger(log.RootLogger(), "ruleaction"))
			pa, err := f(initCtx)
			if err != nil {
				return nil, fmt.Errorf("unable to create rule action service '%s' : %s", serviceCfg.Name, err.Error())
			}
			act = pa
		}

		raService.Act = act
	}

	return raService, nil
}

// SetInput sets input
func (raService *ruleActionService) SetInput(input map[string]interface{}) {
	for k, v := range input {
		raService.Input[k] = v
	}
}

func resolveExpFromTupleScope(tuples map[model.TupleType]model.Tuple, exprs map[string]interface{}) (map[string]interface{}, error) {
	// resolve inputs from tuple scope
	mFactory := mapper.NewFactory(resolve.GetBasicResolver())
	mapper, err := mFactory.NewMapper(exprs)
	if err != nil {
		return nil, err
	}

	tupleScope := make(map[string]interface{})
	for tk, t := range tuples {
		tupleScope[string(tk)] = t.GetMap()
	}

	scope := data.NewSimpleScope(tupleScope, nil)
	return mapper.Apply(scope)
}

// Execute execute rule action service
func (raService *ruleActionService) Execute(ctx context.Context, rs model.RuleSession, rName string, tuples map[model.TupleType]model.Tuple, rCtx model.RuleContext) (done bool, err error) {

	switch raService.Type {

	default:
		return false, fmt.Errorf("unsupported service type - '%s'", raService.Type)

	case config.TypeServiceFunction:
		// invoke function and return, if available
		if raService.Function != nil {
			raService.Function(ctx, rs, rName, tuples, rCtx)
			return true, nil
		}

	case config.TypeServiceActivity:
		// resolve inputs from tuple scope
		resolvedInputs, err := resolveExpFromTupleScope(tuples, raService.Input)
		if err != nil {
			return false, err
		}
		// create activity context and set resolved inputs
		sContext := newServiceContext(raService.Act.Metadata())
		for k, v := range resolvedInputs {
			sContext.SetInput(k, v)
		}
		// run activities Eval
		return raService.Act.Eval(sContext)
	}

	return false, fmt.Errorf("service not executed, something went wrong")
}
