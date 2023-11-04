package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSGreengrassDeviceDefinition struct {
}

func init() {
	i := AWSGreengrassDeviceDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassDeviceDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassDeviceDefinition,
		resource.GreengrassDeviceDefinitionVersion,
	}
}

func (l AWSGreengrassDeviceDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDeviceDefinitions(ctx.Context, &greengrass.ListDeviceDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass device definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(ctx, resource.GreengrassDeviceDefinition, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {

				versions, err := svc.ListDeviceDefinitionVersions(ctx.Context, &greengrass.ListDeviceDefinitionVersionsInput{
					DeviceDefinitionId: v.Id,
					MaxResults:         aws.String("100"),
					NextToken:          nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass device definition versions for %s: %w", *v.Id, err)
				}
				for _, ddId := range versions.Versions {
					dd, err := svc.GetDeviceDefinitionVersion(ctx.Context, &greengrass.GetDeviceDefinitionVersionInput{
						DeviceDefinitionId:        ddId.Id,
						DeviceDefinitionVersionId: ddId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass device definition version for %s, %s: %w", *ddId.Id, *ddId.Version, err)
					}
					ddRes := resource.NewVersion(ctx, resource.GreengrassDeviceDefinitionVersion, dd.Id, dd.Id, dd.Version, dd)
					ddRes.AddRelation(resource.GreengrassDeviceDefinition, v.Id, "")
					// TODO relationships to things
					r.AddRelation(resource.GreengrassDeviceDefinitionVersion, dd.Id, dd.Version)
					rg.AddResource(ddRes)
				}
				return versions.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
