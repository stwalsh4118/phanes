#!/bin/bash
set -e

echo "=== Provisioning Phanes Test VM ==="

# Update package lists
apt-get update -qq

# Install essential packages
apt-get install -y -qq \
    curl \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release

# Install Go 1.21 (or latest stable)
GO_VERSION="1.21.5"
GO_ARCH="amd64"

if [ ! -d "/usr/local/go" ]; then
    echo "Installing Go ${GO_VERSION}..."
    cd /tmp
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o go.tar.gz
    tar -C /usr/local -xzf go.tar.gz
    rm go.tar.gz
fi

# Add Go to PATH for all users
if ! grep -q "/usr/local/go/bin" /etc/profile; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
fi

# Add Go to PATH for current session
export PATH=$PATH:/usr/local/go/bin

# Verify Go installation
go version

# Install system dependencies that modules might need
apt-get install -y -qq \
    ufw \
    fail2ban \
    openssh-server \
    sudo \
    locales

# Ensure SSH is running
systemctl enable sshd || systemctl enable ssh
systemctl start sshd || service ssh start

# Ensure vagrant user exists and has proper setup
if ! id vagrant &>/dev/null; then
    useradd -m -s /bin/bash vagrant
fi

# Ensure vagrant user can use sudo without password (for Vagrant compatibility)
if [ ! -f /etc/sudoers.d/vagrant ]; then
    echo "vagrant ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/vagrant
    chmod 0440 /etc/sudoers.d/vagrant
fi

# Ensure .ssh directory exists for vagrant user
mkdir -p /home/vagrant/.ssh
chown vagrant:vagrant /home/vagrant/.ssh
chmod 700 /home/vagrant/.ssh

# Generate en_US.UTF-8 locale (needed for baseline module)
locale-gen en_US.UTF-8

echo "=== Provisioning complete ==="
echo "Go version: $(go version)"
echo "VM is ready for testing!"


