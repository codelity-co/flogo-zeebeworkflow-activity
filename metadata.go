package zeebeworkflow

import "github.com/project-flogo/core/data/coerce"

type (

	// Settings struct
	Settings struct {
		ZeebeBrokerHost string `md:"zeebeBrokerHost,required"`
		ZeebeBrokerPort int    `md:"zeebeBrokerPort,required"`
		BpmnProcessID   string `md:"bpmnProcessID,required"`
		Command         string `md:"command,required"`
	}

	// Input struct
	Input struct {
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

	return nil
}

func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"zeebeBrokerHost": s.ZeebeBrokerHost,
		"zeebeBrokerPort": s.ZeebeBrokerPort,
		"bpmnProcessID":   s.BpmnProcessID,
		"command":         s.Command,
	}
}

// FromMap method of Input
func (i *Input) FromMap(values map[string]interface{}) error {
	var (
		err   error
		data map[string]interface{}
	)

	data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}

	i.Data = data
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
