package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/glue"

	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGlueWorkflow struct {
}

func init() {
	i := AWSGlueWorkflow{}
	listers = append(listers, i)
}

func (l AWSGlueWorkflow) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueWorkflow,
	}
}

func (l AWSGlueWorkflow) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWorkflows(ctx.Context, &glue.ListWorkflowsInput{
			MaxResults: aws.Int32(25),
		})
		if err != nil {
			return nil, err
		}
		if len(res.Workflows) == 0 {
			return nil, nil
		}
		workflows, err := svc.BatchGetWorkflows(ctx.Context, &glue.BatchGetWorkflowsInput{
			IncludeGraph: aws.Bool(true),
			Names:        res.Workflows,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get glue workflows: %w", err)
		}
		for _, wf := range workflows.Workflows {
			r := resource.New(ctx, resource.GlueWorkflow, wf.Name, wf.Name, wf)
			rg.AddResource(r)
			// TODO: explore nodes/edges
		}
		return res.NextToken, nil
	})
	return rg, err
}
