package acceptance

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDropbear_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDropbearDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDropbearConfigBasic("tf-acc-ssh"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "name", "test-dropbear"),
				),
			},
			{
				ResourceName:      "openwrt_dropbear.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDropbear_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDropbearDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDropbearConfigBasic("tf-acc-ssh"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "name", "test-dropbear"),
				),
			},
			{
				Config: testAccDropbearConfigUpdate("tf-acc-ssh"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "name", "test-dropbear"),
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "password_auth", "false"),
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "root_password_auth", "false"),
					resource.TestCheckResourceAttr("openwrt_dropbear.test", "root_login", "true"),
				),
			},
		},
	})
}

func TestAccSystemLED_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSystemLEDDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemLEDConfigBasic("tf-acc-led", "tp-link:green:power"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_system_led.test", "name", "tf-acc-led"),
					resource.TestCheckResourceAttr("openwrt_system_led.test", "sysfs", "tp-link:green:power"),
				),
			},
			{
				ResourceName:      "openwrt_system_led.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDropbearDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_dropbear" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), parts[0], parts[1])
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return nil
		}
	}

	return nil
}

func testAccCheckSystemLEDDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_system_led" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), parts[0], parts[1])
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return nil
		}
	}

	return nil
}

func testAccDropbearConfigBasic(name string) string {
	return ProviderConfig() + `
resource "openwrt_dropbear" "test" {
  name      = "test-dropbear"
}
`
}

func testAccDropbearConfigUpdate(name string) string {
	return ProviderConfig() + `
resource "openwrt_dropbear" "test" {
  name             = "test-dropbear"
  password_auth    = false
  root_password_auth = false
  root_login       = true
  port             = 22
}
`
}

func testAccSystemLEDConfigBasic(name, sysfs string) string {
	return ProviderConfig() + `
resource "openwrt_system_led" "test" {
  name   = "` + name + `"
  sysfs  = "` + sysfs + `"
}
`
}

func TestAccSystemNTP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSystemNTPDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemNTPConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_system_ntp.test", "enabled", "true"),
				),
			},
			{
				ResourceName:      "openwrt_system_ntp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSystemNTP_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			PreCheck(t)
		},
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSystemNTPDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemNTPConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_system_ntp.test", "enabled", "true"),
				),
			},
			{
				Config: testAccSystemNTPConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openwrt_system_ntp.test", "enabled", "true"),
					resource.TestCheckResourceAttr("openwrt_system_ntp.test", "server.#", "2"),
				),
			},
		},
	})
}

func testAccCheckSystemNTPDestroyed(s *terraform.State) error {
	client := GetTestProvider(&testing.T{})
	if client == nil {
		return nil
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openwrt_system_ntp" {
			continue
		}

		if rs.Primary.ID == "" {
			continue
		}

		parts := splitImportID(rs.Primary.ID)
		if len(parts) != 2 {
			continue
		}

		data, err := client.UCIGetAll(context.Background(), parts[0], parts[1])
		if err != nil {
			return nil
		}

		if len(data) > 0 {
			return nil
		}
	}

	return nil
}

func testAccSystemNTPConfigBasic() string {
	return ProviderConfig() + `
resource "openwrt_system_ntp" "test" {
  name    = "ntp"
  enabled = true
}
`
}

func testAccSystemNTPConfigUpdate() string {
	return ProviderConfig() + `
resource "openwrt_system_ntp" "test" {
  name    = "ntp"
  enabled = true
  server  = ["time.google.com", "time.cloudflare.com"]
}
`
}
