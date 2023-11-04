package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSDocDBParameterGroup struct {
}

func init() {
	i := AWSDocDBParameterGroup{}
	listers = append(listers, i)
}

func (l AWSDocDBParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DocDBParameterGroup}
}

func (l AWSDocDBParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := docdb.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := docdb.NewDescribeDBClusterParameterGroupsPaginator(svc, &docdb.DescribeDBClusterParameterGroupsInput{
		MaxRecords: aws.Int32(100),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, group := range page.DBClusterParameterGroups {
			if !strings.Contains(*group.DBParameterGroupFamily, "docdb") {
				continue
			}
			r := resource.New(ctx, resource.DocDBParameterGroup, group.DBClusterParameterGroupName, group.DBClusterParameterGroupName, group)

			var paramMarker *string
			parameterList := make([]types.Parameter, 0)
			for {
				params, err := svc.DescribeDBClusterParameters(ctx.Context, &docdb.DescribeDBClusterParametersInput{
					DBClusterParameterGroupName: group.DBClusterParameterGroupName,
					Marker:                      paramMarker,
					MaxRecords:                  aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get parameters for %s: %w", *group.DBClusterParameterGroupName, err)
				}
				for _, param := range params.Parameters {
					parameterList = append(parameterList, param)
				}
				if params.Marker == nil {
					break
				}
				paramMarker = params.Marker
			}
			r.AddAttribute("Parameters", parameterList)
			rg.AddResource(r)
		}
	}
	return rg, nil
}
