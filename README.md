Pulumi Libvirt Arch VM Creator
This project uses Pulumi and Go to declaratively provision a local Arch Linux virtual machine on a Linux host using KVM/QEMU and Libvirt.

The main.go file defines a Pulumi program that:

Downloads the latest Arch Linux cloud image.

Creates a cloud-init disk to automatically add your SSH public key and create a sudo user named arch.

Provisions a new KVM domain (a VM) based on the image.

Exports the VM's IP address.

⚠️ Host System Requirements
This guide assumes you are running an Ubuntu/Debian-based host. You will also need:

Go (1.18+) installed.

Pulumi CLI installed and configured.

An SSH key generated at ~/.ssh/id_rsa.pub. (Run ssh-keygen -t rsa -b 4096 if you don't have one).

🚀 Host System Setup (One-Time Only)
Before you can run this project, your host machine must be configured to run KVM/Libvirt.

1. Install Libvirt & KVM Packages
Install the core virtualization packages and genisoimage (which libvirt needs to create cloud-init disks).

Bash

sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients genisoimage
2. Start and Enable Services
Start the libvirtd service and enable it to start on boot.

Bash

sudo systemctl start libvirtd
sudo systemctl enable libvirtd
3. Add Your User to the libvirt Group
To manage VMs as a non-root user, you must be in the libvirt group.

Bash

sudo adduser $USER libvirt
IMPORTANT: You must log out and log back in for this group change to take effect. You can verify it worked by opening a new terminal and running groups. You should see libvirt in the list.

4. Configure the default Storage Pool
Pulumi needs to store the virtual disks in a storage pool. We will create the standard default pool.

Bash

# Define the pool
sudo virsh pool-define-as --name default --type dir --target /var/lib/libvirt/images

# Start the pool
sudo virsh pool-start default

# Set the pool to autostart on boot
sudo virsh pool-autostart default
5. Configure the default Network
The VM needs a network to connect to. We will start the default NAT network.

Bash

# Start the network
sudo virsh net-start default

# Set the network to autostart on boot
sudo virsh net-autostart default
6. Fix AppArmor Permissions
By default, Ubuntu's AppArmor security module will block the VM from reading the disk image we download. We must add an exception.

1. Edit the AppArmor profile:

Bash

sudo nano /etc/apparmor.d/abstractions/libvirt-qemu
2. Add the new line: Scroll to the bottom of the file and add the following line before the final closing }:

  /var/lib/libvirt/images/** rwk,
3. Restart the services: Reload AppArmor and restart libvirtd to apply the changes.

Bash

sudo systemctl restart apparmor.service
sudo systemctl restart libvirtd.service
🏃‍♂️ Project Usage
Once your host is set up, running the project is simple.

1. Set the Libvirt URI: Tell Pulumi to use the "system" (root) libvirt instance.

Bash

pulumi config set libvirt:uri qemu:///system
2. Install Go Dependencies:

Bash

go mod tidy
3. Deploy the VM:

Bash

pulumi up
Pulumi will show you a plan. Type yes to approve it. It will download the Arch image, create the VM, and boot it.

🖥️ Connecting to Your VM
When pulumi up is finished, it will print the VM's IP address in the Outputs section.

Outputs:
    ip: [
        [0]: "192.168.122.239"
    ]
You can now SSH into your new Arch VM as the arch user:

Bash

ssh arch@<YOUR_VM_IP_ADDRESS>
(e.g., ssh arch@192.168.122.239)

Once inside, you can test your sudo access (which is passwordless):

Bash

sudo pacman -Syu
🧹 Cleaning Up
When you are finished with the VM, you can destroy all resources with a single command:

Bash

pulumi destroy