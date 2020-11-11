package zeebeworkflow

import "github.com/project-flogo/core/data/coerce"

type (

	// Settings struct
	Settings struct {
		ZeebeBrokerHost string `md:"zeebeBrokerHost,required"`
		ZeebeBrokerPort int    `md:"zeebeBrokerPort,required"`
		BpmnProcessID   string `md:"bpmnProcessID,required"`
		Command         string `md:"command,required"`
		UsePlainTextConnection bool `md:"usePlainTextConnection"`
		CaCertificatePath string `md:"caCertificatePath"`
	}

	// Input struct
	Input struct {
		MessageName string `md:"messageName,required"`
		MessageCorrelationKey string `md:"messageCorrelationKey,required"`
		Data map[string]interface{} `md:"data,required"`
	}

	// Output struct
	Output struct {
		Status string `md:"status,required"`
		Result interface{} `md:"result,required"`
	}
)

// FromMap method of Settings
func (s *Settings) FromMap(values map[string]interface{}) error {
	var (
		err             error
		zeebeBrokerHost string
		zeebeBrokerPort int
		bpmnProcessID   string
		command         string
		usePlainTextConnection bool
		caCertificatePath string
		
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

	bpmnProcessID, err = coerce.ToString(values["bpmnProcessID"])
	if err != nil {
		return err
	}
	s.BpmnProcessID = bpmnProcessID

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

	caCertificatePath, err = coerce.ToString(values["caCertificatePath"])
	if err != nil {
		return err
	}
	s.CaCertificatePath = caCertificatePath

	return nil
}

// ToMap method of Settings
func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"zeebeBrokerHost": s.ZeebeBrokerHost,
		"zeebeBrokerPort": s.ZeebeBrokerPort,
		"bpmnProcessID":   s.BpmnProcessID,
		"command":         s.Command,
		"usePlainTextConnection": s.UsePlainTextConnection,
		"caCertificatePath": s.CaCertificatePath,
	}
}

// FromMap method of Input
func (i *Input) FromMap(values map[string]interface{}) error {
	var (
		err   error
		messageName string
		messageCorrelationKey string
		data map[string]interface{}
	)

	messageName, err = coerce.ToString(values["messageName"])
	if err != nil {
		return err
	}

	messageCorrelationKey, err = coerce.ToString(values["messageCorrelationKey"])
	if err != nil {
		return err
	}

	data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}

	i.MessageName = messageName
	i.MessageCorrelationKey = messageCorrelationKey
	if data != nil {
		i.Data = data
	}

	return nil
}

// ToMap method of Input
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data": i.Data,
	}
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
