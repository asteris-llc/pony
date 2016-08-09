package templates

import (
	"bytes"
)

type google struct {
}

func google_GenerateTemplate() string {
	g := new(google)

	template := new(bytes.Buffer)

	g.outputProvider(template)

	return template.String()
}

func (g *google) outputProvider(b *bytes.Buffer) {
	b.WriteString(`
{{- /* Required configuration */ -}}
{{- $region := ( variable "region" | required ) -}}
{{- $credentials := ( variable "credentials" | required ) -}}
{{- $project := ( variable "project" | required ) -}}

# Provider information
provider "google" {
  alias = "{{ $region }}"
  region = "{{ $region }}"
  credentials = "${file("{{ $credentials }}")}"
  project = "{{ $project }}"
}
`)
}

var google_template = `
{{- /* Required configuration */ -}}
{{- $region := ( config "region" | env "GOOGLE_REGION" | required "region" ) -}}
{{- $credentials := ( config "credentials" | env "GOOGLE_CREDENTIALS" | env "GOOGLE_CLOUD_KEYFILE_JSON" | env "GCLOUD_KEYFILE_JSON" | required "credentials" ) -}}
{{- $project := ( config "project" | env "GOOGLE_PROJECT" | required "project" ) -}}

{{- /* Optional configuration */ -}}
{{-  $control_type := ( config "instance.control_type" | default "n1-standard-1" ) -}}
{{-  $datacenter := ( config "datacenter" | default "<datacenter-name>" ) -}}
{{-  $gluster_size := ( config "glusterfs.volume_size" | default "100" ) -}}
{{-  $provider := ( print "google." $region ) -}}
{{-  $short_name := ( config "instance.short_name" | default "mantl" ) -}}
{{-  $worker_type := ( config "instance.worker_type" | default "n1-highcpu-2" ) -}}
{{-  $ssh_user := ( config "ssh_user" | default "centos" ) -}}
{{-  $ssh_key := ( config "public_key" | default "~/.ssh/id_rsa.pub" ) -}}

{{- /* resource names */ -}}
{{- $control_t := (print $short_name "-control-%03d") -}}
{{- $resource_t := (print $short_name "-resource-%04d") -}}
{{- $network_name := ( print $short_name "-" $region "-net" ) -}}
{{- $fwext_name := ( print $short_name "-" $region "-firewall-external") -}}
{{- $fwint_name := ( print $short_name "-" $region "-firewall-internal") -}}
{{- $control_instance_t := ( print "google_compute_instance." $control_t ) -}}
{{- $control_address_t := ( print "${" $control_instance_t ".network_interface.0.address}" ) -}}
{{- $resource_instance_t := ( print "google_compute_instance." $resource_t ) -}}
{{- $resource_address_t := ( print "${" $resource_instance_t ".network_interface.0.address}" ) -}}
{{- nodeFormat "control" $control_t -}}
{{- nodeFormat "resource" $resource_t }}

# Provider information
provider "google" {
  alias = "{{ $region }}"
  region = "{{ $region }}"
  credentials = "${file("{{ $credentials }}")}"
  project = "{{ $project }}"
}

resource "google_compute_network" "{{ $network_name }}" {
  provider = "{{ $provider }}"
  name = "{{ $network_name }}"
  ipv4_range = "{{ config "network.ipv4_range" | default "10.0.0.0/16" }}"
} 

# External firewall
resource "google_compute_firewall" "{{ $fwext_name }}" {
  provider = "{{ $provider }}"
  name = "{{ $fwext_name }}"
  network = "${google_compute_network.{{ $network_name }}.name}"
  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports = [
      "22",   # SSH
      "3389", # RDP
      "80",   # HTTP
      "443",  # HTTPs
      "4400", # Chronos
      "4646", # Nomad
      "5050", # Mesos
      "8080", # Marathon
      "18080", # Marathon
      "8500" # Consul UI
    ]
  }
}

# Internal firewall
resource "google_compute_firewall" "{{ $fwint_name }}" {
  provider = "{{ $provider }}"
  name = "{{ $fwint_name }}"
  network = "${google_compute_network.{{ $network_name }}.name}"
  source_ranges = ["${google_compute_network.{{ $network_name }}.ipv4_range}"]

  allow {
    protocol = "tcp"
    ports = ["1-65535"]
  }

  allow {
    protocol = "udp"
    ports = ["1-65535"]
  }
}

{{ range $host := nodes "control" $control_t }}
resource "google_compute_disk" "glusterfs-{{ $host }}" {
  provider = "{{ $provider }}"
  name = "glusterfs-{{ $host }}"
  type = "pd-ssd"
  zone = "{{ $region }}"
  size = "{{ $gluster_size }}"
}

resource "google_compute_instance" "{{ $host }}" {
  provider = "{{ $provider }}"
  name = "{{ $host }}"
  description = "{{ $region }} control node"
  machine_type = "{{ $control_type }}"
  zone = "{{ $region }}"
  can_ip_forward = false
  tags = [ "{{ $short_name }}", "control" ]

  disk {
    image = "centos-7-v20160606"
    auto_delete = true
  }

  disk {
    disk = "${google_compute_disk.glusterfs-{{ $host }}.name}"
    auto_delete = false
    device_name = "glusterfs"
  }

  network_interface {
    network = "${google_compute_network.{{ $network_name }}.name}"
    access_config {}
  }

  metadata {
    dc = "{{ $datacenter }}"
    role = "control"
    ssh_user = "{{ $ssh_user }}"
    sshKeys = "{{ $ssh_user }}:${file("{{ $ssh_key }}")} {{ $ssh_user }}"
    user-data = "${template_cloudinit_config.gce-config.rendered}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo yum -y install cloud-init epel-release",
      "sudo reboot"
    ]

    connection {
      type = "ssh"
      user = "{{ $ssh_user }}"
    }
  }
}
{{ end }}

{{ range $host := nodes "resource" $resource_t }}
resource "google_compute_instance" "{{ $host }}" {
  provider = "{{ $provider }}"
  name = "{{ $host }}"
  description = "{{ $region }} resource node"
  machine_type = "{{ $worker_type }}"
  zone = "{{ $region }}"
  can_ip_forward = false
  tags = [ "{{ $short_name }}", "worker" ]

  disk {
    image = "centos-7-v20160606"
    auto_delete = true
  }

  network_interface {
    network = "${google_compute_network.{{ $network_name }}.name}"
    access_config {}
  }

  metadata {
    dc = "{{ $datacenter }}"
    role = "resource"
    ssh_user = "{{ $ssh_user }}"
    sshKeys = "{{ $ssh_user }}:${file("{{ $ssh_key }}")} {{ $ssh_user }}"
    user-data = "${template_cloudinit_config.gce-config.rendered}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo yum -y install cloud-init epel-release",
      "sudo reboot"
    ]

    connection {
      type = "ssh"
      user = "{{ $ssh_user }}"
    }
  }
}
{{ end }}

{{ if gt (len (nodes "control" $control_t )) 0 }}
resource "template_cloudinit_config" "gce-config" {
  gzip = false
  base64_encode = false

  part {
    content_type = "text/cloud-config"
    content = <<EOT
#cloud-config
yum_repos:
  bintray-chrisaubuchon:
    name: Chris Aubuchon - BinTray
    baseurl: https://dl.bintray.com/chrisaubuchon/rpm
    enabled: true
    gpgcheck: false

packages:
  - mantl-bootstrap
  - consul-cli
  - consul-mantl
  - consul-ui
  - smlr
  - jq
  - mantl-dns

runcmd:
  - [ ln, -sf, /etc/localtime, /usr/share/zoneinfo/Etc/UTC ]
  - [ /usr/bin/mantl-bootstrap, listen ]
  - [ update-ca-trust ]
  - [ systemctl, enable, consul ]
  - [ yum, update, -y ]
  - [ reboot ]

users:
  - name: chris
    primary-group: wheel
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh-authorized-keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDKptmKAMk+5tKYCmfz7VD1YXzBSjnjcvc3BKEhbXqhgOPdIvyTgMd6N95BukadWbVfHfy9NGIBB/36b/gP1l9Vw7DTW7RLnyVOjTitbHG87EUppjB7N1YL0sgJ3Fr0y7ON10lIdZBT2vkyAsXmFwShnRRvkFLr5l4LaYAI2pnx/C612s2MwGbZORpOVY23ozKdaQjP70K2AuXKQmFFKnofKiBLFNheE5r8yjbPqqDhpLAC8kCIG0Gosd5uNuk1N2Y/QjXivbDWMVsnCECQN+NVWYR2/4m5BFDW5DHbMzLGxbttWF65jgZnM0CuD+pln8WwYc1QPTKFMz8LE9luZwe1 Chris.Aubuchon@gmail.com
EOT
  }
}

resource "null_resource" "bootstrap-leader" {
  depends_on = [ {{ join "," (nodes "control" (printf "%q" $control_instance_t)) (nodes "resource" (printf "%q" $resource_instance_t))}} ]
  triggers {
    control_nodes = "{{ join "," (nodes "control" $control_instance_t) }}"
    resource_nodes = "{{ join "," (nodes "resource" $resource_instance_t) }}"
  }

  connection = {
    type = "ssh"
    user = "{{ $ssh_user }}"
    host = "${google_compute_instance.{{ bootstrapNode "control" }}.network_interface.0.access_config.0.assigned_nat_ip}"
  }

  provisioner "remote-exec" {
    inline = [
      "while [ ! -f /usr/bin/mantl-bootstrap ]; do echo waiting for cloud-init; sleep 10; done",
      "/usr/bin/mantl-bootstrap bootstrap --servers={{ join "," (nodes "control" $control_address_t) }}{{if gt (len (nodes "resource" $resource_t)) 0 }} --clients={{ join "," (nodes "resource" $resource_address_t) }}{{ end -}}"
    ]
  }
}
{{ end }}
`
