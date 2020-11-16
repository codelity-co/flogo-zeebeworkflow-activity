package zeebeworkflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

type TokenProvider struct {
	token string
}

func (tp *TokenProvider) ApplyCredentials(ctx context.Context, headers map[string]string) error {
	headers["Authorization"] = tp.token
	return nil
}

func (tp *TokenProvider) ShouldRetryRequest(ctx context.Context, err error) bool {
	return status.Code(err) == codes.DeadlineExceeded
}

// New function is factory method of activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	var (
		err          error
		zeebeClient  zbc.Client
		clientConfig *zbc.ClientConfig
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
	clientConfig = &zbc.ClientConfig{
		GatewayAddress:         fmt.Sprintf("%v:%v", s.ZeebeBrokerHost, s.ZeebeBrokerPort),
		UsePlaintextConnection: s.UsePlainTextConnection,
	}

	zeebeClient, err = zbc.NewClient(clientConfig)
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
	case "CreateWorkflowInstance":
		result["createWorkflowInstanceResponse"], err = a.createWorkflowInstance(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "CancelWorkflowInstance":
		result["cancelWorkflowInstanceResponse"], err = a.cancelWorkflowInstance(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "PublishMessage":
		result["publishMessageResponse"], err = a.publishMessage(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "CompleteJob":
		result["completeJobResponse"], err = a.completeJob(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "FailJob":
		result["failJobResponse"], err = a.failJob(ctx, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "ResolveIncident":
		result["resolveIncidentResponse"], err = a.resolveIncident(ctx, input.Data)
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
		err           error
		bpmnProcessID string
		data          map[string]interface{}
		request       commands.CreateInstanceCommandStep3
	)

	ctx.Logger().Debug("Running createWorkflowInstance func...")
	ctx.Logger().Debugf("input: %v", input)

	ctx.Logger().Debug("Extracting bpmnProcessID")
	bpmnProcessID, err = coerce.ToString(input["bpmnProcessID"])
	if err != nil {
		ctx.Logger().Errorf("Get messageName error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting data")
	data, err = coerce.ToObject(input["data"])
	if err != nil {
		ctx.Logger().Errorf("Get messageName error: %v", err)
		return nil, err
	}

	if data != nil {
		request, err = a.zeebeClient.NewCreateInstanceCommand().BPMNProcessId(bpmnProcessID).LatestVersion().VariablesFromMap(data)
		if err != nil {
			ctx.Logger().Errorf("Failed to prepare create workflow instance request: %v", err)
			return nil, err
		}
	} else {
		request = a.zeebeClient.NewCreateInstanceCommand().BPMNProcessId(bpmnProcessID).LatestVersion()
	}

	response, err := request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send create workflow instance request: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting response")
	result := map[string]interface{}{
		"bpmnProcessID":                response.GetBpmnProcessId(),
		"version":                      response.GetVersion(),
		"workflowKey":                  response.GetWorkflowKey(),
		"workflowInstanceKey":          response.GetWorkflowInstanceKey(),
		"createWorkflowInstanceStatus": true,
	}

	ctx.Logger().Debug("Finished createWorkflowInstance func successfully")
	return result, nil
}

func (a *Activity) cancelWorkflowInstance(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var (
		err                 error
		workflowInstanceKey int64
	)

	ctx.Logger().Debug("Running cancelWorkflowInstance func...")
	ctx.Logger().Debugf("input: %v", input)

	ctx.Logger().Debug("Extracting workflowInstanceKey")
	workflowInstanceKey, err = coerce.ToInt64(input["workflowInstanceKey"])
	if err != nil {
		ctx.Logger().Errorf("Get workfowInstanceKey error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Creating request")
	request := a.zeebeClient.NewCancelInstanceCommand().WorkflowInstanceKey(workflowInstanceKey)

	ctx.Logger().Debug("Sending request")
	_, err = request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send cancel workflow instance request: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"workflowInstanceKey":          workflowInstanceKey,
		"cancelWorkflowInstanceStatus": true,
	}

	ctx.Logger().Debug("Finished createWorkflowInstance func successfully")
	return result, nil
}

func (a *Activity) publishMessage(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var (
		err                   error
		messageName           string
		messageCorrelationKey string
		messageTtlToLive      time.Duration
		messageData           map[string]interface{}
	)

	ctx.Logger().Debug("Running publish message func...")
	ctx.Logger().Debugf("input: %v", input)

	ctx.Logger().Debug("Extracting messageName")
	messageName, err = coerce.ToString(input["messageName"])
	if err != nil {
		ctx.Logger().Errorf("Get messageName error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting messageCorrelationKey")
	messageCorrelationKey, err = coerce.ToString(input["messageCorrelationKey"])
	if err != nil {
		ctx.Logger().Errorf("Get messageCorrelationKey error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting ttlToLiveDuration")
	durationString, err := coerce.ToString(input["messageTtlToLive"])
	if err != nil {
		ctx.Logger().Errorf("Get ttlToLiveDuration error: %v", err)
		return nil, err
	}
	messageTtlToLive, err = time.ParseDuration(durationString)
	if err != nil {
		ctx.Logger().Errorf("Get ttlToLiveDuration error: %v", err)
		return nil, err
	}

	if _, exists := input["data"]; exists {
		ctx.Logger().Debug("Extracting data")
		messageData, err = coerce.ToObject(input["data"])
		if err != nil {
			ctx.Logger().Errorf("Get data error: %v", err)
			return nil, err
		}
	}

	ctx.Logger().Debug("Creating request")
	request := a.zeebeClient.NewPublishMessageCommand().MessageName(messageName).CorrelationKey(messageCorrelationKey).TimeToLive(messageTtlToLive)
	if messageData != nil {
		request, err = request.VariablesFromMap(messageData)
		if err != nil {
			ctx.Logger().Errorf("Publish Message request preparatioon error: %v", err)
			return nil, err
		}
	}

	ctx.Logger().Debug("Sending request")
	_, err = request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send publish message request: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"messageName":           messageName,
		"messageCorrelationKey": messageCorrelationKey,
		"messageTtlToLive":      messageTtlToLive,
		"publishMessageStatus":  true,
	}

	ctx.Logger().Debug("Finished publishMessage func successfully")
	return result, nil
}

func (a *Activity) resolveIncident(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {
	var (
		err         error
		incidentKey int64
	)

	ctx.Logger().Debug("Running resolve workflow instance func...")

	ctx.Logger().Debug("Extracting incidentKey")
	if incidentKey, err = coerce.ToInt64(input["incidentKey"]); err != nil {
		ctx.Logger().Errorf("Get incidentKey error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Creating request")
	request := a.zeebeClient.NewResolveIncidentCommand().IncidentKey(incidentKey)

	ctx.Logger().Debug("Sending request")
	if response, err := request.Send(context.Background()); err != nil {
		ctx.Logger().Errorf("Failed to send resolve incident request: %v", err)
		return nil, err
	} else {
		result := map[string]interface{}{
			"incidentKey":                 incidentKey,
			"resolveIncidentResponseText": response.String(),
		}
		return result, nil
	}
}

func (a *Activity) completeJob(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {

	var (
		err      error
		jobKey   int64
		data     map[string]interface{}
		response *pb.CompleteJobResponse
	)

	ctx.Logger().Debug("Running complete Job func...")

	ctx.Logger().Debug("Extracting jobKey")
	if input["jobKey"] == nil {
		err = errors.New("missing jobKey")
		ctx.Logger().Errorf("Get joyKey error: %v", err)
		return nil, err
	}
	jobKey, err = coerce.ToInt64(input["jobKey"])
	if err != nil {
		ctx.Logger().Errorf("Get joyKey error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Extracting data")
	if input["data"] != nil {
		data, err = coerce.ToObject(input["data"])
		if err != nil {
			ctx.Logger().Errorf("Get data error: %v", err)
			return nil, err
		}
	}

	ctx.Logger().Debug("Creating request")

	request, err := a.zeebeClient.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(data)
	if err != nil {
		ctx.Logger().Errorf("Complete job request preparatioon error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Sending request")

	response, err = request.Send(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send complete job request: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"jobKey":                  jobKey,
		"completeJobResponseText": response.String(),
	}
	return result, nil
}

func (a *Activity) failJob(ctx activity.Context, input map[string]interface{}) (map[string]interface{}, error) {

	var (
		err      error
		jobKey   int64
		retries  int32
		response *pb.FailJobResponse
	)

	ctx.Logger().Debug("Running fail Job func...")

	ctx.Logger().Debug("Extracting jobKey")
	if input["jobKey"] == nil {
		err = errors.New("missing jobKey")
		ctx.Logger().Errorf("Get joyKey error: %v", err)
		return nil, err
	}
	jobKey, err = coerce.ToInt64(input["jobKey"])
	if err != nil {
		ctx.Logger().Errorf("Get joyKey error: %v", err)
		return nil, err
	}

	ctx.Logger().Debug("Creating request")
	request := a.zeebeClient.NewFailJobCommand().JobKey(jobKey).Retries(a.activitySettings.FailJobRetries)

	ctx.Logger().Debug("Sending request")
	if response, err = request.Send(context.Background()); err != nil {
		ctx.Logger().Errorf("Failed to send fail job request: %v", err)
		return nil, err
	} else {
		result := map[string]interface{}{
			"jobKey":              jobKey,
			"retries":             retries,
			"failJobResponseText": response.String(),
		}
		return result, nil
	}

}
