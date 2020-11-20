package zeebeworkflow

import "github.com/project-flogo/core/data/coerce"

// Settings struct
type Settings struct {
	ZeebeBrokerHost        string `md:"zeebeBrokerHost,required"`
	ZeebeBrokerPort        int    `md:"zeebeBrokerPort,required"`
	Command                string `md:"command,required"`
	UsePlainTextConnection bool   `md:"usePlainTextConnection"`

	FailJobRetries int32 `md:"failJobRetries"`
}

// FromMap method of Settings
func (s *Settings) FromMap(values map[string]interface{}) error {
	var (
		err                    error
		zeebeBrokerHost        string
		zeebeBrokerPort        int
		command                string
		usePlainTextConnection bool
		failJobRetries         int32
	)

	zeebeBrokerHost, err = coerce.ToString(values["zeebeBrokerHost"])
	if err != nil {
		return err
	}
	s.ZeebeBrokerHost = zeebeBrokerHost

	zeebeBrokerPort, err = coerce.ToInt(values["zeebeBrokerPort"])
	if err != nil {
		return err
	}
	s.ZeebeBrokerPort = zeebeBrokerPort

	command, err = coerce.ToString(values["command"])
	if err != nil {
		return err
	}
	s.Command = command

	usePlainTextConnection, err = coerce.ToBool(values["usePlainTextConnection"])
	if err != nil {
		return err
	}
	s.UsePlainTextConnection = usePlainTextConnection

	failJobRetries, err = coerce.ToInt32(values["failJobRetries"])
	if err != nil {
		return err
	}
	s.FailJobRetries = failJobRetries

	return nil
}

// ToMap method of Settings
func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"zeebeBrokerHost":        s.ZeebeBrokerHost,
		"zeebeBrokerPort":        s.ZeebeBrokerPort,
		"command":                s.Command,
		"usePlainTextConnection": s.UsePlainTextConnection,
		"failJobRetries":         s.FailJobRetries,
	}
}

// Input struct
type Input struct {
	BpmnProcessID         string                 `md:"bpmnProcessID,required"`
	WorkflowInstanceKey   int64                  `md:"workflowInstanceKey"`
	MessageName           string                 `md:"messageName"`
	MessageCorrelationKey string                 `md:"messageCorrelationKey"`
	MessageTtlToLiveString string                `md:"messageTtlToLiveString"`
	IncidentKey           int64                  `md:"incidentKey"`
	JobKey                int64                  `md:"jobKey"`
	Data                  map[string]interface{} `md:"data"`
}

// FromMap method of Input
func (i *Input) FromMap(values map[string]interface{}) error {
	var (
		err                   error
		bpmnProcessID         string
		workflowInstanceKey   int64
		messageName           string
		messageCorrelationKey string
		messageTtlToLiveString      string
		incidentKey           int64
		jobKey                int64
		data                  map[string]interface{}
	)

	bpmnProcessID, err = coerce.ToString(values["bpmnProcessID"])
	if err != nil {
		return err
	}

	workflowInstanceKey, err = coerce.ToInt64(values["workflowInstanceKey"])
	if err != nil {
		return err
	}

	messageName, err = coerce.ToString(values["messageName"])
	if err != nil {
		return err
	}

	messageCorrelationKey, err = coerce.ToString(values["messageCorrelationKey"])
	if err != nil {
		return err
	}

	messageTtlToLiveString, err = coerce.ToString(values["messageTtlToLiveString"])
	if err != nil {
		return err
	}

	incidentKey, err = coerce.ToInt64(values["incidentKey"])
	if err != nil {
		return err
	}

	jobKey, err = coerce.ToInt64(values["jobKey"])
	if err != nil {
		return err
	}

	data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}

	if bpmnProcessID != "" {
		i.BpmnProcessID = bpmnProcessID
	}

	if workflowInstanceKey > 0 {
		i.WorkflowInstanceKey = workflowInstanceKey
	}
	if messageName != "" {
		i.MessageName = messageName
	}
	if messageCorrelationKey != "" {
		i.MessageCorrelationKey = messageCorrelationKey
	}
	if messageTtlToLiveString != "" {
		i.MessageTtlToLiveString = messageTtlToLiveString
	}
	if incidentKey > 0 {
		i.IncidentKey = incidentKey
	}
	if jobKey > 0 {
		i.JobKey = jobKey
	}
	if data != nil {
		i.Data = data
	}

	return nil
}

// ToMap method of Input
func (i *Input) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	if i.BpmnProcessID != "" {
		result["bpmnProcessID"] = i.BpmnProcessID
	}
	if i.WorkflowInstanceKey > 0 {
		result["workflowInstanceKey"] = i.WorkflowInstanceKey
	}
	if i.MessageName != "" {
		result["messageName"] = i.MessageName
	}
	if i.MessageCorrelationKey != "" {
		result["messageCorrelationKey"] = i.MessageCorrelationKey
	}
	if i.MessageTtlToLiveString != "" {
		result["messageTtlToLiveString"] = i.MessageTtlToLiveString
	}
	if i.IncidentKey > 0 {
		result["incidentKey"] = i.IncidentKey
	}
	if i.JobKey > 0 {
		result["jobKey"] = i.JobKey
	}
	if i.Data != nil {
		result["data"] = i.Data
	}
	return result
}

// Output struct
type Output struct {
	Status string      `md:"status,required"`
	Result interface{} `md:"result,required"`
}

// FromMap of Output
func (o *Output) FromMap(values map[string]interface{}) error {
	var (
		err    error
		status string
		result interface{}
	)

	status, err = coerce.ToString(values["status"])
	if err != nil {
		return err
	}
	o.Status = status

	result, err = coerce.ToAny(values["result"])
	if err != nil {
		return err
	}
	o.Result = result

	return nil
}

// ToMap method of Output
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"status": o.Status,
		"result": o.Result,
	}
}
