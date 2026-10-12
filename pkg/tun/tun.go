// Package tun provides TUN device management for Linux.
package tun

import (
	"fmt"
	"net"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	cIFF_TUN   = 0x0001
	cIFF_NO_PI = 0x1000

	defaultMTU = 1500
)

// ifReq is the ifreq struct for ioctl calls.
type ifReq struct {
	Name  [unix.IFNAMSIZ]byte
	Flags uint16
	_     [22]byte
}

// Device represents an opened TUN device.
type Device struct {
	file      *os.File
	name      string
	mtu       int
	ipv4CIDR  string
	ipv6CIDR  string
}

// Config holds TUN device creation parameters.
type Config struct {
	Name     string
	IPv4CIDR string
	IPv6CIDR string
	MTU      int
}

// CreateTUN creates and configures a new TUN device.
func CreateTUN(cfg Config) (*Device, error) {
	if cfg.MTU <= 0 {
		cfg.MTU = defaultMTU
	}

	// Open /dev/net/tun
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("opening /dev/net/tun: %w", err)
	}

	// Configure the TUN device
	var req ifReq
	req.Flags = cIFF_TUN | cIFF_NO_PI
	copy(req.Name[:], cfg.Name)

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		unix.Close(fd)
		return nil, fmt.Errorf("TUNSETIFF ioctl: %w", errno)
	}

	devName := string(req.Name[:unix.IFNAMSIZ])
	// Trim null bytes
	for i, c := range devName {
		if c == 0 {
			devName = devName[:i]
			break
		}
	}

	file := os.NewFile(uintptr(fd), "/dev/net/tun")

	dev := &Device{
		file:     file,
		name:     devName,
		mtu:      cfg.MTU,
		ipv4CIDR: cfg.IPv4CIDR,
		ipv6CIDR: cfg.IPv6CIDR,
	}

	// Bring up the interface and set addresses
	if err := dev.configure(); err != nil {
		file.Close()
		return nil, fmt.Errorf("configuring TUN device: %w", err)
	}

	return dev, nil
}

// configure sets up the TUN interface with IP addresses and MTU.
func (d *Device) configure() error {
	link, err := net.InterfaceByName(d.name)
	if err != nil {
		// Try using ip command as fallback
		return d.configureViaIPCmd()
	}
	_ = link
	return d.configureViaIPCmd()
}

// configureViaIPCmd uses the `ip` command to configure the interface.
func (d *Device) configureViaIPCmd() error {
	// Set link up
	if err := runCmd("ip", "link", "set", d.name, "up"); err != nil {
		return fmt.Errorf("setting link up: %w", err)
	}

	// Set MTU
	if err := runCmd("ip", "link", "set", d.name, "mtu", fmt.Sprintf("%d", d.mtu)); err != nil {
		return fmt.Errorf("setting MTU: %w", err)
	}

	// Add IPv4 address
	if d.ipv4CIDR != "" {
		if err := runCmd("ip", "addr", "add", d.ipv4CIDR, "dev", d.name); err != nil {
			return fmt.Errorf("adding IPv4 address: %w", err)
		}
	}

	// Add IPv6 address
	if d.ipv6CIDR != "" {
		if err := runCmd("ip", "addr", "add", d.ipv6CIDR, "dev", d.name); err != nil {
			return fmt.Errorf("adding IPv6 address: %w", err)
		}
	}

	return nil
}

// Read reads a packet from the TUN device.
func (d *Device) Read(buf []byte) (int, error) {
	return d.file.Read(buf)
}

// Write writes a packet to the TUN device.
func (d *Device) Write(buf []byte) (int, error) {
	return d.file.Write(buf)
}

// Close closes the TUN device.
func (d *Device) Close() error {
	return d.file.Close()
}

// Name returns the TUN device name.
func (d *Device) Name() string {
	return d.name
}

// MTU returns the configured MTU.
func (d *Device) MTU() int {
	return d.mtu
}

// File returns the underlying file descriptor for use with poll/epoll.
func (d *Device) File() *os.File {
	return d.file
}