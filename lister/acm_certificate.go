package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSAcmCertificate struct {
}

func init() {
	i := AWSAcmCertificate{}
	listers = append(listers, i)
}

func (l AWSAcmCertificate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AcmCertificate}
}

func (l AWSAcmCertificate) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := acm.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := acm.NewListCertificatesPaginator(svc, &acm.ListCertificatesInput{
		MaxItems: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, cert := range page.CertificateSummaryList {

			res, err := svc.DescribeCertificate(ctx.Context, &acm.DescribeCertificateInput{CertificateArn: cert.CertificateArn})
			if err != nil {
				return nil, fmt.Errorf("unable to describe certificate %s: %w", *cert.CertificateArn, err)
			}
			//if arn.IsArnP(res.Certificate.CertificateArn) {
			certArn := arn.ParseP(res.Certificate.CertificateArn)
			r := resource.New(ctx, resource.AcmCertificate, certArn.ResourceId, certArn.ResourceId, res.Certificate)
			//}
			tagRes, err := svc.ListTagsForCertificate(ctx.Context, &acm.ListTagsForCertificateInput{
				CertificateArn: cert.CertificateArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list tags for cert %s: %w", *cert.CertificateArn, err)
			}
			for _, tag := range tagRes.Tags {
				r.Tags[*tag.Key] = *tag.Value
			}
			rg.AddResource(r)
		}
	}
	return rg, nil
}
