package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSCognitoIdentityPool struct {
}

func init() {
	i := AWSCognitoIdentityPool{}
	listers = append(listers, i)
}

func (l AWSCognitoIdentityPool) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CognitoIdentityPool,
	}
}

func (l AWSCognitoIdentityPool) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cognitoidentity.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListIdentityPools(ctx.Context, &cognitoidentity.ListIdentityPoolsInput{
			MaxResults: 60,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list identity pools: %w", err)
		}
		for _, identityPool := range res.IdentityPools {
			pool, err := svc.DescribeIdentityPool(ctx.Context, &cognitoidentity.DescribeIdentityPoolInput{
				IdentityPoolId: identityPool.IdentityPoolId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe identity pool %s: %w", *identityPool.IdentityPoolName, err)
			}
			r := resource.New(ctx, resource.CognitoIdentityPool, pool.IdentityPoolId, pool.IdentityPoolName, pool)

			roles, err := svc.GetIdentityPoolRoles(ctx.Context, &cognitoidentity.GetIdentityPoolRolesInput{
				IdentityPoolId: identityPool.IdentityPoolId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe roles for identity pool %s: %w", *identityPool.IdentityPoolId, err)
			}
			r.AddAttribute("RoleAttachment", roles)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
