package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSGreengrassCoreDefinition struct {
}

func init() {
	i := AWSGreengrassCoreDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassCoreDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassCoreDefinition,
		resource.GreengrassCoreDefinitionVersion,
	}
}

func (l AWSGreengrassCoreDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListCoreDefinitions(ctx.Context, &greengrass.ListCoreDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass core definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(ctx, resource.GreengrassCoreDefinition, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListCoreDefinitionVersions(ctx.Context, &greengrass.ListCoreDefinitionVersionsInput{
					CoreDefinitionId: v.Id,
					MaxResults:       aws.String("100"),
					NextToken:        nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass core definition versions for %s: %w", *v.Id, err)
				}
				for _, cdId := range versions.Versions {
					cd, err := svc.GetCoreDefinitionVersion(ctx.Context, &greengrass.GetCoreDefinitionVersionInput{
						CoreDefinitionId:        cdId.Id,
						CoreDefinitionVersionId: cdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass core definition version for %s, %s: %w", *cdId.Id, *cdId.Version, err)
					}
					cdRes := resource.NewVersion(ctx, resource.GreengrassCoreDefinitionVersion, cd.Id, cd.Id, cd.Version, cd)
					cdRes.AddRelation(resource.GreengrassCoreDefinition, v.Id, "")
					// TODO relationships to things
					r.AddRelation(resource.GreengrassCoreDefinitionVersion, cd.Id, cd.Version)
					rg.AddResource(cdRes)
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
