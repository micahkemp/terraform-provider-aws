package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccDataSourceAwsEc2CoipPools_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2CoipPoolsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "pool_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2CoipPools_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2CoipPoolsConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "pool_ids.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2CoipPoolsConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}
`
}

func testAccDataSourceAwsEc2CoipPoolsConfigFilter() string {
	return `
data "aws_ec2_coip_pools" "all" {}

data "aws_ec2_coip_pools" "test" {
  filter {
    name   = "coip-pool.pool-id"
    values = [tolist(data.aws_ec2_coip_pools.all.pool_ids)[0]]
  }
}
`
}
