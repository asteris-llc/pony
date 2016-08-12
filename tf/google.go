package tf

var google_template = `
module "gce" { source = "github.com/ChrisAubuchon/pony-config/gce/base" }
`

var not_used = `
variable credentials {
  description = "Path to the JSON file for GCE credentials"
  default = "account.json"
}
variable project {
  description = "The ID of the project to apply any resources to"
}
variable region {
  description = "The GCE region to operate under"
}
variable short_name {
  default = "mantl"
  description = "Prefix for created resources"
}

variable "meta_required_variables" {
  type = "list"

  default = [
    "credentials",
    "project",
    "region",
    "short_name",
  ]
}

provider "google" {
  credentials = "${file(var.credentials)}"
  project = "${var.project}"
  region = "${var.region}"
}

module "gce-network" {
  description = "Google network"
  source = "github.com/ChrisAubuchon/pony-config/gce/network"
}
`
