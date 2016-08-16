# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = '2'

@script = <<SCRIPT
GOROOT="/opt/go"
GOPATH="/opt/gopath"

# Get the ARCH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

# Install tools
sudo yum install -y git

# Install Go
cd /tmp
curl -q https://storage.googleapis.com/golang/go1.6.3.linux-${ARCH}.tar.gz -o /tmp/go.tar.gz
tar -xvf go.tar.gz
sudo mv go $GOROOT
sudo chmod 0775 $GOROOT
sudo chown vagrant:vagrant $GOROOT
rm -f /tmp/go.tar.gz

# Setup the GOPATH
sudo mkdir -p ${GOPATH}
sudo mkdir -p ${GOPATH}/bin
sudo chown -R vagrant:vagrant ${GOPATH}
sudo chmod 0775 ${GOPATH}
sudo chmod 0775 ${GOPATH}/bin
cat <<EOF >/tmp/gopath.sh
export GOPATH="${GOPATH}"
export GOROOT="${GOROOT}"
export PATH="/opt/go/bin:\\\$GOPATH/bin:\\\$PATH"
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh

# Install glide
curl https://glide.sh/get | sh

# Install go tools
go get github.com/mitchellh/gox

cd ${GOPATH}/src/github.com/mitchellh/gox
go build
cp gox ${GOPATH}/bin

cd ${GOPATH}/src/github.com/asteris-lc/pony
make vendor
make
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.provision 'shell', inline: @script
  config.vm.synced_folder '.', '/opt/gopath/src/github.com/asteris-llc/pony', type: "rsync", rsync__exclude: "vendor/"

  %w[vmware_fusion vmware_workstation].each do |_|
    config.vm.provider 'p' do |v|
      v.vmx['memsize'] = '2048'
      v.vmx['numvcpus'] = '2'
      v.vmx['cpuid.coresPerSocket'] = '1'
    end
  end

  config.vm.define '64bit' do |n1|
    n1.vm.box = 'centos/7'
  end
end
