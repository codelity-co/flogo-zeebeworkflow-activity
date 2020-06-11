package zeebeworkflow

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

type ZeebeworkflowActivityTestSuite struct {
	suite.Suite
}

func TestZeebeworkflowActivityTestSuite(t *testing.T) {
	suite.Run(t, new(ZeebeworkflowActivityTestSuite))
}

func (suite *ZeebeworkflowActivityTestSuite) SetupSuite() {
	command := exec.Command("docker", "start", "zeebe")
	err := command.Run()
	if err != nil {
		fmt.Println(err.Error())
		command := exec.Command("docker", "run", "-p", "26500-26502:26500-26502", "--name", "zeebe", "-d", "camunda/zeebe:latest")
		err := command.Run()
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		time.Sleep(10 * time.Second)
	}

}

func (suite *ZeebeworkflowActivityTestSuite) BeforeTest(suiteName, testName string) {

	switch testName {
	case "TestZeebeworkflowActivity_CreateWorkflowInstance":

		zeebeClient, err := zbc.NewClient(&zbc.ClientConfig{
			GatewayAddress:         "localhost:26500",
			UsePlaintextConnection: true,
		})
		if err != nil {
			panic(err)
		}
		response, err := zeebeClient.NewDeployWorkflowCommand().AddResourceFile("./test/order-process.bpmn").Send(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("response text: %v", response.String()))
	}

}

func (suite *ZeebeworkflowActivityTestSuite) TestZeebeworkflowActivity_Register() {

	ref := activity.GetRef(&Activity{})
	act := activity.Get(ref)

	assert.NotNil(suite.T(), act)
}

func (suite *ZeebeworkflowActivityTestSuite) TestZeebeworkflowActivity_Settings() {
	t := suite.T()

	settings := &Settings{}

	iCtx := test.NewActivityInitContext(settings, nil)
	_, err := New(iCtx)
	assert.Nil(t, err)
}

func (suite *ZeebeworkflowActivityTestSuite) TestZeebeworkflowActivity_CreateWorkflowInstance() {
	t := suite.T()

	settings := &Settings{
		ZeebeBrokerHost: "127.0.0.1",
		ZeebeBrokerPort: 26500,
		BpmnProcessID:   "order-process",
		Command:         "Create",
		UsePlainTextConnection: true,
	}

	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("in", nil)
	_, err = act.Eval(tc)
	assert.Nil(t, err)

	status := tc.GetOutput("status").(string)
	assert.Equal(t, status, "SUCCESS", "Status must be SUCCESS")
	result := tc.GetOutput("result").(map[string]interface{})
	assert.NotNil(t, result["createWorkflowInstanceResponse"])
}
