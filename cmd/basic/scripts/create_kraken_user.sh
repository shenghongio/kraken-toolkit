#!/bin/bash
set -e
id {{.Username}} >/dev/null 2>&1 && echo "SKIP" && exit 0
useradd -m {{.Username}}
mkdir -p ~{{.Username}}/.ssh
cat >> ~{{.Username}}/.ssh/authorized_keys << 'EOF'
{{.PublicKey}}
EOF
chmod 600 ~{{.Username}}/.ssh/authorized_keys
chmod 700 ~{{.Username}}/.ssh
chown -R {{.Username}}:{{.Username}} ~{{.Username}}/.ssh
echo "{{.Username}} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/{{.Username}}
echo "CREATED"