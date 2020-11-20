package zeebeworkflow

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/project-flogo/core/activity"
	"github.com/zeebe-io/zeebe/clients/go/pkg/commands"
	"github.com/zeebe-io/zeebe/clients/go/pkg/pb"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

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

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

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
	logger.Infof("ctx.Settings(): %v", ctx.Settings())
	err = s.FromMap(ctx.Settings())
	if err != nil {
		return nil, err
	}
	logger.Infof("Settings: %v", s)

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

	logger.Debugf("Activity Settings: %v", a.activitySettings)

	input := &Input{Logger: logger}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}
	logger.Debugf("Input: %v", input)

	output := &Output{}
	result := make(map[string]interface{})

	logger.Debugf("Command: %v", a.activitySettings.Command)
	switch a.activitySettings.Command {
	case "CreateWorkflowInstance":
		result["createWorkflowInstanceResponse"], err = a.createWorkflowInstance(ctx, input.BpmnProcessID, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "CancelWorkflowInstance":
		result["cancelWorkflowInstanceResponse"], err = a.cancelWorkflowInstance(ctx, input.WorkflowInstanceKey)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "PublishMessage":
		result["publishMessageResponse"], err = a.publishMessage(ctx, input.MessageName, input.MessageCorrelationKey, input.MessageTtlToLiveString, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "CompleteJob":
		result["completeJobResponse"], err = a.completeJob(ctx, input.JobKey, input.Data)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "FailJob":
		result["failJobResponse"], err = a.failJob(ctx, input.JobKey)
		if err != nil {
			output.Status = "ERROR"
			output.Result = err.Error()
			_ = ctx.SetOutputObject(output)
			return true, err
		}
	case "ResolveIncident":
		result["resolveIncidentResponse"], err = a.resolveIncident(ctx, input.IncidentKey)
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

func (a *Activity) createWorkflowInstance(ctx activity.Context, bpmnProcessID string, data map[string]interface{}) (map[string]interface{}, error) {
	var (
		err     error
		request commands.CreateInstanceCommandStep3
	)

	ctx.Logger().Debug("Running createWorkflowInstance func...")
	ctx.Logger().Debugf("bpmnProcessID: %v", bpmnProcessID)
	ctx.Logger().Debugf("data: %v", data)

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

func (a *Activity) cancelWorkflowInstance(ctx activity.Context, workflowInstanceKey int64) (map[string]interface{}, error) {
	var (
		err error
	)

	ctx.Logger().Debug("Running cancelWorkflowInstance func...")
	ctx.Logger().Debugf("workflowInstanceKey: %v", workflowInstanceKey)

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

func (a *Activity) publishMessage(ctx activity.Context, messageName string, messageCorrelationKey string, messageTtlToLiveString string, data map[string]interface{}) (map[string]interface{}, error) {
	var (
		err              error
		messageTtlToLive time.Duration
	)

	ctx.Logger().Debug("Running publish message func...")
	ctx.Logger().Debugf("messageName: %v", messageName)
	ctx.Logger().Debugf("messageCorrelationKey: %v", messageCorrelationKey)
	ctx.Logger().Debugf("messageTtlToLiveString: %v", messageTtlToLiveString)
	ctx.Logger().Debugf("data: %v", data)

	if messageTtlToLiveString != "" {
		messageTtlToLive, err = time.ParseDuration(messageTtlToLiveString)
		if err != nil {
			ctx.Logger().Errorf("Get ttlToLiveDurationString error: %v", err)
			return nil, err
		}
	}

	ctx.Logger().Debug("Creating request")
	request := a.zeebeClient.NewPublishMessageCommand().MessageName(messageName).CorrelationKey(messageCorrelationKey).TimeToLive(messageTtlToLive)
	if data != nil {
		request, err = request.VariablesFromMap(data)
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
		"messageName":            messageName,
		"messageCorrelationKey":  messageCorrelationKey,
		"messageTtlToLiveString": messageTtlToLiveString,
		"publishMessageStatus":   true,
	}

	ctx.Logger().Debug("Finished publishMessage func successfully")
	return result, nil
}

func (a *Activity) resolveIncident(ctx activity.Context, incidentKey int64) (map[string]interface{}, error) {

	ctx.Logger().Debug("Running resolve workflow instance func...")
	ctx.Logger().Debugf("incidentKey: %v", incidentKey)

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

func (a *Activity) completeJob(ctx activity.Context, jobKey int64, data map[string]interface{}) (map[string]interface{}, error) {

	var (
		err      error
		response *pb.CompleteJobResponse
	)

	ctx.Logger().Debug("Running complete Job func...")
	ctx.Logger().Debugf("jobKey: %v", jobKey)
	ctx.Logger().Debugf("data: %v", data)

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

func (a *Activity) failJob(ctx activity.Context, jobKey int64) (map[string]interface{}, error) {

	var (
		err      error
		retries  int32
		response *pb.FailJobResponse
	)

	ctx.Logger().Debug("Running fail Job func...")
	ctx.Logger().Debugf("jobKey: %v", jobKey)

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
