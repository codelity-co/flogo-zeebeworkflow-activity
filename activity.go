package zeebeworkflow

import (
	"context"
	"fmt"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/zeebe-io/zeebe/clients/go/pkg/commands"
	"github.com/zeebe-io/zeebe/clients/go/pkg/pb"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

func init() {
	_ = activity.Register(&Activity{}) //activity.Register(&Activity{}, New) to create instances using factory method 'New'
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// New function is factory method of activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	var (
		err         error
		zeebeClient zbc.Client
	)
	logger := ctx.Logger()

	// Activity settings
	s := &Settings{}
	err = s.FromMap(ctx.Settings())
	if err != nil {
		return nil, err
	}
	logger.Debugf("Settings: %v", s)

	// Connect to Zeebe broker
	zeebeClient, err = zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         fmt.Sprintf("%v:%v", s.ZeebeBrokerHost, s.ZeebeBrokerPort),
		UsePlaintextConnection: s.UsePlainTextConnection,
	})
	if err != nil {
		logger.Errorf("Zeebe broker connection error: %v", err)
		return nil, err
	}

	// Create Activity
	act := &Activity{
		activityInitContext: ctx,
		activitySettings:    s,
		zeebeClient:         zeebeClient,
	}

	return act, nil
}

// Activity struct
type Activity struct {
	activityInitContext activity.InitContext
	activitySettings    *Settings
	zeebeClient         zbc.Client
}

// Metadata method of Activity returns activity metdata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval method of activity
func (a *Activity) Eval(ctx activity.Context) (bool, error) {
	var (
		err error
	)

	logger := ctx.Logger()

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	logger.Debugf("Input: %v", input)

	output := &Output{}
	result := make(map[string]interface{})

	switch a.activitySettings.Command {
	case "Create":
		result["createWorkflowInstanceResponse"], err = a.createWorkflowInstance(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "Cancel":
		result["cancelWorkflowInstanceResponse"], err = a.cancelWorkflowInstance(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	default:
		err = fmt.Errorf("Invalid Zeebe workflow instance command")
		output.Status = "ERROR"
		output.Result = err.Error()
		_ = ctx.SetOutputObject(output)
		return true, err
	}

	output.Status = "SUCCESS"
	output.Result = result
	logger.Debugf("Output: %v", input)

	err = ctx.SetOutputObject(output)
	if err != nil {
		logger.Errorf("Failed to set output object in context: %v", err)
		return true, err
	}

	return true, nil
}

// Cleanup method of activity
func (a *Activity) Cleanup() error {
	var err error
	logger := a.activityInitContext.Logger()

	// Close Zeebe broker connection
	err = a.zeebeClient.Close()
	if err != nil {
		logger.Errorf("Failed to close Zeebe broker connection: %v", err)
		return err
	}

	return nil
}

func (a *Activity) createWorkflowInstance(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var (
		err      error
		request  commands.CreateInstanceCommandStep3
		response *pb.CreateWorkflowInstanceResponse
	)

	ctx.Logger().Debug("Running createWorkflowInstance func...")
	ctx.Logger().Debugf("input: %v", input)

	if input != nil {
		request, err = a.zeebeClient.NewCreateInstanceCommand().BPMNProcessId(a.activitySettings.BpmnProcessID).LatestVersion().VariablesFromMap(input)
	} else {
		request = a.zeebeClient.NewCreateInstanceCommand().BPMNProcessId(a.activitySettings.BpmnProcessID).LatestVersion()
	}

	if err != nil {
		ctx.Logger().Errorf("Failed to prepare create workflow instance request: %v", err)
		return nil, err
	}
	response, err = request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send create workflow instance request: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting response")
	result := map[string]interface{}{
		"bpmnProcessId":       response.GetBpmnProcessId(),
		"version":             response.GetVersion(),
		"workflowKey":         response.GetWorkflowKey(),
		"workflowInstanceKey": response.GetWorkflowInstanceKey(),
	}

	ctx.Logger().Debug("Finished createWorkflowInstance func successfully")
	return result, nil
}

func (a *Activity) cancelWorkflowInstance(ctx activity.Context, input map[string]interface{}) (string, error) {
	var (
		err                 error
		workflowInstanceKey int64
		request             commands.DispatchCancelWorkflowInstanceCommand
		response            *pb.CancelWorkflowInstanceResponse
	)

	ctx.Logger().Debug("Running cancelWorkflowInstance func...")
	ctx.Logger().Debugf("input: %v", input)

	ctx.Logger().Debug("Extracting workflowInstanceKey")
	workflowInstanceKey, err = coerce.ToInt64(input["workflowInstanceKey"])
	if err != nil {
		ctx.Logger().Errorf("Get workfowInstanceKey error: %v", err)
		return "", err
	}

	ctx.Logger().Debug("Creating request")
	request = a.zeebeClient.NewCancelInstanceCommand().WorkflowInstanceKey(workflowInstanceKey)
	ctx.Logger().Debug("Sending request")
	response, err = request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send cancel workflow instance request: %v", err)
		return "", err
	}

	ctx.Logger().Debug("Finished createWorkflowInstance func successfully")
	return response.String(), nil
}
