package app

import (
	"context"
	"errors"
	"github.com/gogo/protobuf/types"
	"github.com/temporalio/tcld/protogen/api/cloud/cloudservice/v1"
	"github.com/temporalio/tcld/protogen/api/cloud/nexus/v1"
	"github.com/temporalio/tcld/protogen/api/cloud/operation/v1"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	cloudservicemock "github.com/temporalio/tcld/protogen/apimock/cloudservice/v1"
	"github.com/urfave/cli/v2"
)

func TestNexus(t *testing.T) {
	suite.Run(t, new(NexusTestSuite))
}

type NexusTestSuite struct {
	suite.Suite
	cliApp           *cli.App
	mockCtrl         *gomock.Controller
	mockCloudService *cloudservicemock.MockCloudServiceClient
}

func (s *NexusTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockCloudService = cloudservicemock.NewMockCloudServiceClient(s.mockCtrl)
	out, err := NewNexusCommand(func(ctx *cli.Context) (*NexusClient, error) {
		return &NexusClient{
			ctx:    context.TODO(),
			client: s.mockCloudService,
		}, nil
	})
	s.Require().NoError(err)
	AutoConfirmFlag.Value = true
	s.cliApp = &cli.App{
		Name:     "test",
		Commands: []*cli.Command{out.Command},
		Flags: []cli.Flag{
			AutoConfirmFlag,
		},
	}
}

func (s *NexusTestSuite) RunCmd(args ...string) error {
	return s.cliApp.Run(append([]string{"tcld"}, args...))
}

func (s *NexusTestSuite) AfterTest(suiteName, testName string) {
	s.mockCtrl.Finish()
}

func getExampleNexusEndpoint() *nexus.Endpoint {
	return &nexus.Endpoint{
		Id:              "test-endpoint-id",
		ResourceVersion: "test-resource-version",
		Spec: &nexus.EndpointSpec{
			TargetSpec: &nexus.EndpointTargetSpec{
				WorkerTargetSpec: &nexus.WorkerTargetSpec{
					NamespaceId: "test-namespace-name.test-account-id",
					TaskQueue:   "test-task-queue",
				},
			},
			Name: "test_name",
			PolicySpecs: []*nexus.EndpointPolicySpec{
				{
					AllowedCloudNamespacePolicySpec: &nexus.AllowedCloudNamespacePolicySpec{
						NamespaceId: "test-caller-namespace.test-account-id",
					},
				},
				{
					AllowedCloudNamespacePolicySpec: &nexus.AllowedCloudNamespacePolicySpec{
						NamespaceId: "test-caller-namespace-2.test-account-id",
					},
				},
			},
		},
		State:            "activating",
		AsyncOperationId: "test-request-id",
		CreatedTime:      &types.Timestamp{Seconds: time.Date(time.Now().Year(), time.April, 12, 0, 0, 0, 0, time.UTC).Unix()},
		LastModifiedTime: &types.Timestamp{Seconds: time.Date(time.Now().Year(), time.April, 14, 0, 0, 0, 0, time.UTC).Unix()},
	}
}

func (s *NexusTestSuite) TestEndpointGet() {
	s.Error(s.RunCmd("nexus", "endpoint", "get"))

	s.Error(s.RunCmd("nexus", "endpoint", "get", "--name"))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found")).Times(1)
	s.Error(s.RunCmd("nexus", "endpoint", "get", "--name", "test-nexus-endpoint-name"))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "get", "--name", "test-nexus-endpoint-name"))
}

func (s *NexusTestSuite) TestEndpointList() {
	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(nil, errors.New("list error")).Times(1)
	s.Error(s.RunCmd("nexus", "endpoint", "list"))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "list"))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "list"))
}

func (s *NexusTestSuite) TestEndpointCreate() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().CreateNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("create error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "create",
		"--name", exampleEndpoint.Spec.Name,
		"--target-namespace", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.NamespaceId,
		"--target-task-queue", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.TaskQueue,
		"--allow-namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--allow-namespace", exampleEndpoint.Spec.PolicySpecs[1].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "create error")

	s.mockCloudService.EXPECT().CreateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.CreateNexusEndpointResponse{
		EndpointId: exampleEndpoint.Id,
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "create",
		"--name", exampleEndpoint.Spec.Name,
		"--target-namespace", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.NamespaceId,
		"--target-task-queue", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.TaskQueue,
		"--allow-namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--allow-namespace", exampleEndpoint.Spec.PolicySpecs[1].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	))
}

func (s *NexusTestSuite) TestEndpointUpdate() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "update",
		"--name", exampleEndpoint.Spec.Name,
		"--target-task-queue", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.TaskQueue+"-updated",
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("update error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "update",
		"--name", exampleEndpoint.Spec.Name,
		"--target-task-queue", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.TaskQueue+"-updated",
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "update error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.UpdateNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "update",
		"--name", exampleEndpoint.Spec.Name,
		"--target-task-queue", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.TaskQueue+"-updated",
		"--request-id", exampleEndpoint.AsyncOperationId,
	))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.UpdateNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "update",
		"--name", exampleEndpoint.Spec.Name,
		"--target-namespace", exampleEndpoint.Spec.TargetSpec.WorkerTargetSpec.NamespaceId+"-updated",
		"--request-id", exampleEndpoint.AsyncOperationId,
	))

	s.EqualError(s.RunCmd("nexus", "endpoint", "update",
		"--name", exampleEndpoint.Spec.Name,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "no updates to be made")
}

func (s *NexusTestSuite) TestEndpointAllowedNamespaceAdd() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "add",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", "test-another-caller-namespace.test-account-id",
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("update error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "add",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", "test-another-caller-namespace.test-account-id",
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "update error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.UpdateNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "add",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", "test-another-caller-namespace.test-account-id",
		"--request-id", exampleEndpoint.AsyncOperationId,
	))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "add",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "no updates to be made")
}

func (s *NexusTestSuite) TestEndpointAllowedNamespaceSet() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "set",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("update error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "set",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "update error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.UpdateNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "set",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	))
}

func (s *NexusTestSuite) TestEndpointAllowedNamespaceList() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "list", "--name", exampleEndpoint.Spec.Name), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{}, errors.New("list error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "list", "--name", exampleEndpoint.Spec.Name), "list error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "list", "--name", exampleEndpoint.Spec.Name))
}

func (s *NexusTestSuite) TestEndpointAllowedNamespaceRemove() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "remove",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("update error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "remove",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "update error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().UpdateNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.UpdateNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "remove",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", exampleEndpoint.Spec.PolicySpecs[0].AllowedCloudNamespacePolicySpec.NamespaceId,
		"--request-id", exampleEndpoint.AsyncOperationId,
	))

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "allowed-namespace", "remove",
		"--name", exampleEndpoint.Spec.Name,
		"--namespace", "test-another-caller-namespace.test-account-id",
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "no updates to be made")
}

func (s *NexusTestSuite) TestEndpointDelete() {
	exampleEndpoint := getExampleNexusEndpoint()

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{}}, nil).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "delete",
		"--name", exampleEndpoint.Spec.Name,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "endpoint not found")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().DeleteNexusEndpoint(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error")).Times(1)
	s.EqualError(s.RunCmd("nexus", "endpoint", "delete",
		"--name", exampleEndpoint.Spec.Name,
		"--request-id", exampleEndpoint.AsyncOperationId,
	), "delete error")

	s.mockCloudService.EXPECT().GetNexusEndpoints(gomock.Any(), gomock.Any()).Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexus.Endpoint{getExampleNexusEndpoint()}}, nil).Times(1)
	s.mockCloudService.EXPECT().DeleteNexusEndpoint(gomock.Any(), gomock.Any()).Return(&cloudservice.DeleteNexusEndpointResponse{
		AsyncOperation: &operation.AsyncOperation{
			Id: exampleEndpoint.AsyncOperationId,
		},
	}, nil).Times(1)
	s.NoError(s.RunCmd("nexus", "endpoint", "delete",
		"--name", exampleEndpoint.Spec.Name,
		"--request-id", exampleEndpoint.AsyncOperationId,
	))
}
