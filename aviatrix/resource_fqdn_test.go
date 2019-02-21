package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/AviatrixSystems/go-aviatrix/goaviatrix"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccFQDN_basic(t *testing.T) {
	var fqdn goaviatrix.FQDN
	rName := fmt.Sprintf("%s", acctest.RandString(5))

	skipAcc := os.Getenv("SKIP_FQDN")
	if skipAcc == "yes" {
		t.Skip("Skipping FQDN test as SKIP_FQDN is set")
	}

	preGatewayCheck(t, ". Set SKIP_FQDN to yes to skip FQDN tests")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFQDNDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFQDNConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFQDNExists("aviatrix_fqdn.foo", &fqdn),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "fqdn_tag",
						fmt.Sprintf("tff-%s", rName)),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "fqdn_status", "enabled"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "fqdn_mode", "white"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "gw_list.#", "1"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "gw_list.0",
						fmt.Sprintf("tfg-%s", rName)),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "domain_names.#", "1"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "domain_names.0.fqdn",
						"facebook.com"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "domain_names.0.proto", "tcp"),
					resource.TestCheckResourceAttr("aviatrix_fqdn.foo", "domain_names.0.port", "443"),
				),
			},
		},
	})
}

func testAccFQDNConfigBasic(rName string) string {
	return fmt.Sprintf(`

resource "aviatrix_account" "test" {
    account_name = "tfa-%s"
	cloud_type = 1
	aws_account_number = "%s"
	aws_iam = "false"
	aws_access_key = "%s"
	aws_secret_key = "%s"
}

resource "aviatrix_gateway" "test" {
	cloud_type = 1
	account_name = "${aviatrix_account.test.account_name}"
	gw_name = "tfg-%[1]s"
	vpc_id = "%[5]s"
	vpc_reg = "%[6]s"
	vpc_size = "t2.micro"
	vpc_net = "%[7]s"
}

resource "aviatrix_fqdn" "foo" {
	fqdn_tag = "tff-%[1]s"
	fqdn_status = "enabled"
	fqdn_mode = "white"
	gw_list = ["${aviatrix_gateway.test.gw_name}"]
	domain_names =  [
		{
			fqdn = "facebook.com"
			proto = "tcp"
			port = "443"
		}
	]
}
`, rName, os.Getenv("AWS_ACCOUNT_NUMBER"), os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"),
		os.Getenv("AWS_VPC_ID"), os.Getenv("AWS_REGION"), os.Getenv("AWS_VPC_NET"))
}

func testAccCheckFQDNExists(n string, fqdn *goaviatrix.FQDN) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("FQDN Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no FQDN ID is set")
		}

		client := testAccProvider.Meta().(*goaviatrix.Client)

		foundFQDN := &goaviatrix.FQDN{
			FQDNTag: rs.Primary.Attributes["fqdn_tag"],
		}

		_, err := client.GetFQDNTag(foundFQDN)

		if err != nil {
			return err
		}

		if foundFQDN.FQDNTag != rs.Primary.ID {
			return fmt.Errorf("FQDN not found")
		}

		*fqdn = *foundFQDN

		return nil
	}
}

func testAccCheckFQDNDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*goaviatrix.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aviatrix_fqdn" {
			continue
		}
		foundFQDN := &goaviatrix.FQDN{
			FQDNTag: rs.Primary.Attributes["fqdn_tag"],
		}
		_, err := client.GetFQDNTag(foundFQDN)

		if err != goaviatrix.ErrNotFound {
			return fmt.Errorf("FQDN still exists")
		}
	}
	return nil
}
