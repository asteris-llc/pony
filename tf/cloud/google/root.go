package google

var root_module = `
variable credentials {
	type = "string"
	description = "Path to the JSON file for GCE credentials"
	default = "account.json"
}

variable project {
	type = "string"
	description = "The ID of the project to apply any resources to"
}

variable region {
	type = "string"
	description = "The GCE region to operate under"
	default = "us-central1"
}

variable zones {
	type = "string"
	description = "Comma separated list of GCE regions"
	default = "us-central1-a,us-central1-b"
}

variable short_name {
	type = "string"
	description = "The prefix for created resources"
	default = "mantl"
}

variable ssh_user {
	type = "string"
	description = "The user to use for the connection"
	default = "centos"
}

variable ssh_key {
	type = "string"
	description = "The SSH public key to add to the instances"
	default = "~/.ssh/id_rsa.pub"
}

variable image {
	type = "string"
	decription = "Machine image to use"
	default = "centos-7-v20160606"
}

variable datacenter {
	type = "string"
	desciption = "Consul datacenter name"
	default = "gce"
}

variable "meta_required_variables" {
	type = "list"

	default = [
		"short_name"
	]
}

variable "meta_provider_variables" {
	type = "list"

	default = [
		"credentials",
		"project",
		"region",
		"zones"
	]
}

variable "meta_destroy_variables" {
	type = "list"

	default = [
		"credentials",
		"project",
		"region"
	]
}

provider "google" {
	credentials = "${file(var.credentials)}"
	project = "${var.project}"
	region = "${var.region}"
}

module "gce-network" {
	description = "Google network"
	source = "builtin:network"
}

module "control-nodes" {
	description = "Mantl Control Nodes"
	source = "builtin:instance"
	count = "3"
	machine_type = "n1-standard-1"
	network_name = "${module.gce-network.network_name}"
	volume_type = "pd-ssd"
	role = "control"
}

module "edge-nodes" {
	description = "Mantl Edge Nodes"
	source = "builtin:instance"
	count = "1"
	machine_type = "n1-standard-1"
	network_name = "${module.gce-network.network_name}"
	role = "edge"
}

module "worker-nodes" {
	description = "Mantl Worker Nodes"
	source = "builtin:instance"
	count = "2"
	machine_type = "n1-standard-2"
	network_name = "${module.gce-network.network_name}"
	role = "worker"
}

output "cloud" {
	value = "google"
}

output "credentials" {
	value = "${var.credentials}"
}

output "project" {
	value = "${var.project}"
}

output "region" {
	value = "${var.region}"
}
`
