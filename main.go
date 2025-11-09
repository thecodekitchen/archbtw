package main

import (
	"fmt"
	"os"

	// 1. Using the correct, official public package
	"github.com/pulumi/pulumi-libvirt/sdk/go/libvirt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// --- 1. Define Your Cloud-Init Configuration ---
		// !! IMPORTANT: Replace with your public SSH key !!
		publicKey, err := os.ReadFile("/home/atd/.ssh/id_rsa.pub")
		if err != nil {
			// As a fallback, paste your public key string directly
			// publicKey = []byte("ssh-rsa AAAA...")
			return fmt.Errorf("could not read public key: %w. "+
				"Please ensure ~/.ssh/id_rsa.pub exists, or paste your key in main.go", err)
		}

		// cloud-init user-data
		userData := pulumi.String(fmt.Sprintf(`#cloud-config
users:
  - name: arch
    ssh_authorized_keys:
      - %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: [wheel]
    shell: /bin/bash
`, string(publicKey)))

		// 2. Corrected function name 'NewCloudInitDisk'
		cloudInitDisk, err := libvirt.NewCloudInitDisk(ctx, "arch-init", &libvirt.CloudInitDiskArgs{
			UserData: userData,
		})
		if err != nil {
			return err
		}

		// --- 2. Define the Base Arch Linux Image ---
		baseVolume, err := libvirt.NewVolume(ctx, "arch-base-qcow2", &libvirt.VolumeArgs{
			Source: pulumi.String("https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2"),
			Format: pulumi.String("qcow2"),
		})
		if err != nil {
			return err
		}

		// --- 3. Create the VM's Main Disk ---
		mainDisk, err := libvirt.NewVolume(ctx, "arch-disk", &libvirt.VolumeArgs{
			BaseVolumeId: baseVolume.ID(),
			Size:         pulumi.Int(10 * 1024 * 1024 * 1024), // 10 GB
			Format:       pulumi.String("qcow2"),
		})
		if err != nil {
			return err
		}

		// --- 4. Define and Create the Virtual Machine ---
		// 3. Capturing the VM resource in 'archVM'
		archVM, err := libvirt.NewDomain(ctx, "arch-vm", &libvirt.DomainArgs{
			Name:   pulumi.String("arch-vm"),
			Memory: pulumi.Int(2048), // 2 GB RAM
			Vcpu:   pulumi.Int(2),

			// 4. Using the correct 'CloudinitId' property
			Cloudinit: cloudInitDisk.ID(),

			// Attach the main disk
			Disks: libvirt.DomainDiskArray{
				&libvirt.DomainDiskArgs{
					VolumeId: mainDisk.ID(),
				},
			},

			// Configure the network to use the 'default' libvirt NAT network
			NetworkInterfaces: libvirt.DomainNetworkInterfaceArray{
				&libvirt.DomainNetworkInterfaceArgs{
					NetworkName:  pulumi.String("default"),
					WaitForLease: pulumi.Bool(true),
				},
			},

			// Use the base image's console
			Consoles: libvirt.DomainConsoleArray{
				&libvirt.DomainConsoleArgs{
					Type:       pulumi.String("pty"),
					TargetPort: pulumi.String("0"),
					TargetType: pulumi.String("serial"),
				},
			},
			Graphics: &libvirt.DomainGraphicsArgs{
				Type:       pulumi.String("spice"),
				ListenType: pulumi.String("address"),
				Autoport:   pulumi.Bool(true),
			},
		})
		if err != nil {
			return err
		}

		// --- 5. (Optional) Export the VM's IP Address ---
		// 5. Correctly exporting the IP from the 'archVM' variable
		ctx.Export("ip", archVM.NetworkInterfaces.ApplyT(
			func(interfaces []libvirt.DomainNetworkInterface) []string {
				var ips []string
				for _, iface := range interfaces {
					if iface.Addresses != nil {
						ips = append(ips, iface.Addresses...)
					}
				}
				return ips
			},
		))

		return nil
	})
}
