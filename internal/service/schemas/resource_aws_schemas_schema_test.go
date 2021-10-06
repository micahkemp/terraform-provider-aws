package schemas_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/schemas"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
)

const (
	testAccAWSSchemasSchemaContent = `
{
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "title": "Event"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Event": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          }
        }
      }
    }
  }
}
`

	testAccAWSSchemasSchemaContentUpdated = `
{
  "openapi": "3.0.0",
  "info": {
    "version": "2.0.0",
    "title": "Event"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Event": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          }
        }
	  }
	}
  }
}
`
)

func TestAccAWSSchemasSchema_basic(t *testing.T) {
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "schemas", fmt.Sprintf("schema/%s/%s", rName, rName)),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "OpenApi3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "version_created_date"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSchemasSchema_disappears(t *testing.T) {
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceSchema(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSchemasSchema_ContentDescription(t *testing.T) {
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfigContentDescription(rName, testAccAWSSchemasSchemaContent, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "content", testAccAWSSchemasSchemaContent),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasSchemaConfigContentDescription(rName, testAccAWSSchemasSchemaContentUpdated, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "content", testAccAWSSchemasSchemaContentUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
			{
				Config: testAccAWSSchemasSchemaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
				),
			},
		},
	})
}

func TestAccAWSSchemasSchema_Tags(t *testing.T) {
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, schemas.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSchemasSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSchemasSchemaConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSchemasSchemaConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSchemasSchemaConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemasSchemaExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSchemasSchemaDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_schema" {
			continue
		}

		name, registryName, err := tfschemas.SchemaParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfschemas.FindSchemaByNameAndRegistryName(conn, name, registryName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge Schemas Schema %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSchemasSchemaExists(n string, v *schemas.DescribeSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Schema ID is set")
		}

		name, registryName, err := tfschemas.SchemaParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn

		output, err := tfschemas.FindSchemaByNameAndRegistryName(conn, name, registryName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSSchemasSchemaConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q
}
`, rName, testAccAWSSchemasSchemaContent)
}

func testAccAWSSchemasSchemaConfigContentDescription(rName, content, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q
  description   = %[3]q
}
`, rName, content, description)
}

func testAccAWSSchemasSchemaConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, testAccAWSSchemasSchemaContent, tagKey1, tagValue1)
}

func testAccAWSSchemasSchemaConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`	
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, testAccAWSSchemasSchemaContent, tagKey1, tagValue1, tagKey2, tagValue2)
}
