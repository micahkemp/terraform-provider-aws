package cur_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cur "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcur "github.com/hashicorp/terraform-provider-aws/internal/service/cur"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cur_report_definition", &resource.Sweeper{
		Name: "aws_cur_report_definition",
		F:    testSweepCurReportDefinitions,
	})
}

func testSweepCurReportDefinitions(region string) error {
	c, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	client := c.(*conns.AWSClient)
	if !testAccRegionSupportsCur(client.Region, client.Partition) {
		log.Printf("[WARN] Skipping Cost and Usage Report Definitions sweep for %s: not supported in this region", region)
		return nil
	}

	conn := client.CURConn

	input := &cur.DescribeReportDefinitionsInput{}
	var sweeperErrs *multierror.Error
	err = conn.DescribeReportDefinitionsPages(input, func(page *cur.DescribeReportDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportDefinition := range page.ReportDefinitions {
			r := tfcur.ResourceReportDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(reportDefinition.ReportName))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Cost And Usage Report Definitions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Cost And Usage Report Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsCurReportDefinition_basic(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "arn", "cur", fmt.Sprintf("definition/%s", reportName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "arn", "cur", fmt.Sprintf("definition/%s", reportName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func testAccAwsCurReportDefinition_textOrCsv(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
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

func testAccAwsCurReportDefinition_parquet(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
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

func testAccAwsCurReportDefinition_athena(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := "data"
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{"ATHENA"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
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

func testAccAwsCurReportDefinition_refresh(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := true
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "true"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
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

func testAccAwsCurReportDefinition_overwrite(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
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

func testAccAwsCurReportDefinition_disappears(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCur(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcur.ResourceReportDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsCurReportDefinitionDestroy(s *terraform.State) error {
	conn := testAccProviderCur.Meta().(*conns.AWSClient).CURConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cur_report_definition" {
			continue
		}

		matchingReportDefinition, err := tfcur.FindReportDefinitionByName(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Report Definition (%s): %w", rs.Primary.ID, err)
		}

		if matchingReportDefinition == nil {
			continue
		}

		return fmt.Errorf("Report Definition still exists: %q", rs.Primary.ID)
	}
	return nil

}

func testAccCheckAwsCurReportDefinitionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderCur.Meta().(*conns.AWSClient).CURConn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		matchingReportDefinition, err := tfcur.FindReportDefinitionByName(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Report Definition (%s): %w", rs.Primary.ID, err)
		}

		if matchingReportDefinition == nil {
			return fmt.Errorf("Report Definition does not exist: %q", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName, prefix string) string {
	return acctest.ConfigCompose(
		testAccCurRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  acl           = "private"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[3]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
`, reportName, bucketName, prefix))
}

func testAccAwsCurReportDefinitionConfig_additional(reportName string, bucketName string, bucketPrefix string, format string, compression string, additionalArtifacts []string, refreshClosedReports bool, reportVersioning string) string {
	artifactsStr := strings.Join(additionalArtifacts, "\", \"")

	if len(additionalArtifacts) > 0 {
		artifactsStr = fmt.Sprintf("additional_artifacts       = [\"%s\"]", artifactsStr)
	} else {
		artifactsStr = ""
	}

	return acctest.ConfigCompose(
		testAccCurRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[2]s"
  acl           = "private"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = "%[1]s"
  time_unit                  = "DAILY"
  format                     = "%[4]s"
  compression                = "%[5]s"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = "%[3]s"
  s3_region                  = aws_s3_bucket.test.region
	%[6]s
  refresh_closed_reports = %[7]t
  report_versioning      = "%[8]s"
}
`, reportName, bucketName, bucketPrefix, format, compression, artifactsStr, refreshClosedReports, reportVersioning))
}

func TestCheckDefinitionPropertyCombination(t *testing.T) {
	type propertyCombinationTestCase struct {
		additionalArtifacts []string
		compression         string
		format              string
		prefix              string
		reportVersioning    string
		shouldError         bool
	}

	testCases := map[string]propertyCombinationTestCase{
		"TestAthenaAndAdditionalArtifacts": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndEmptyPrefix": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndOverrideReport": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaWithNonParquetFormat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestParquetFormatWithoutParquetCompression": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestCSVFormatWithParquetCompression": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestRedshiftWithParquetformat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestQuicksightWithParquetformat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaValidCombination": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestMultipleArtifacts": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
	}

	for name, tCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tfcur.CheckReportDefinitionPropertyCombination(
				tCase.additionalArtifacts,
				tCase.compression,
				tCase.format,
				tCase.prefix,
				tCase.reportVersioning,
			)

			if tCase.shouldError && err == nil {
				t.Fail()
			} else if !tCase.shouldError && err != nil {
				t.Fail()
			}
		})
	}
}
