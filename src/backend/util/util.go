package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
)

var (
	// AuthTokenHeaderName http header for token transport
	AuthTokenHeaderName = "x-wg-gen-plus-auth"
	// RegexpEmail check valid email
	RegexpEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// ReadFile file content
func ReadFile(path string) (bytes []byte, err error) {
	bytes, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// WriteFile content to file
func WriteFile(path string, bytes []byte) (err error) {
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// FileExists check if file exists
func FileExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists check if directory exists
func DirectoryExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetAvailableIp search for an available ip in cidr against a list of reserved ips
func GetAvailableIp(cidr string, reserved []string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	// this two addresses are not usable
	broadcastAddr := BroadcastAddr(ipnet).String()
	networkAddr := ipnet.IP.String()

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); Inc(ip) {
		ok := true
		address := ip.String()
		for _, r := range reserved {
			if address == r {
				ok = false
				break
			}
		}
		if ok && address != networkAddr && address != broadcastAddr {
			return address, nil
		}
	}

	return "", errors.New("no more available address from cidr")
}

// IsIPv6 check if given ip is IPv6
func IsIPv6(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() == nil
}

// IsValidIp check if ip is valid
func IsValidIp(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidCidr check if CIDR is valid
func IsValidCidr(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// GetIpFromCidr get ip from cidr
func GetIpFromCidr(cidr string) (string, error) {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

// Increments the IP address by 1
func Inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// BroadcastAddr returns the last address in the given network, or the broadcast address.
func BroadcastAddr(n *net.IPNet) net.IP {
	// The golang net package doesn't make it easy to calculate the broadcast address. :(
	var broadcast net.IP
	if len(n.IP) == 4 {
		broadcast = net.ParseIP("0.0.0.0").To4()
	} else {
		broadcast = net.ParseIP("::")
	}
	for i := 0; i < len(n.IP); i++ {
		broadcast[i] = n.IP[i] | ^n.Mask[i]
	}
	return broadcast
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateRandomULASubnet generates a random IPv6 Unique Local Address subnet (/64)
// as defined in RFC 4193. Returns the subnet in CIDR notation.
func GenerateRandomULASubnet() (string, error) {
	// Allocate 5 bytes (40 bits) for the Global ID
	globalID, err := GenerateRandomBytes(5)
	if err != nil {
		return "", err
	}

	// Allocate 2 bytes (16 bits) for the Subnet ID
	subnetID, err := GenerateRandomBytes(2)
	if err != nil {
		return "", err
	}

	// Format: FD + 5-byte Global ID + 2-byte Subnet ID + 8 bytes of zeros (for the interface ID)
	// FD = prefix FD00::/8 with L bit set to 1 (unique local address)
	ipv6 := net.IP{
		0xFD,
		globalID[0],
		globalID[1],
		globalID[2],
		globalID[3],
		globalID[4],
		subnetID[0],
		subnetID[1],
		0, 0, 0, 0, 0, 0, 0, 0, // Interface ID (64 bits) - all zeros
	}

	// Return the subnet with /64 prefix length
	return ipv6.String() + "/64", nil
}

// Generate a random IPv4 /24 subnet in the 10.0.0.0/8 range.
// Avoids 10.0.0.0/24 which is often used for LANs.
func GenerateRandomIPv4Subnet() (string, error) {
	// Generate 2 random bytes for the second and third octets
	randomBytes, err := GenerateRandomBytes(2)
	if err != nil {
		return "", err
	}

	// Extract the values from the random bytes
	secondOctet := int(randomBytes[0])
	thirdOctet := int(randomBytes[1])

	// Special case: avoid 10.0.0.0/24
	if secondOctet == 0 && thirdOctet == 0 {
		thirdOctet = 1
	}

	// Format the subnet as a CIDR string
	subnet := fmt.Sprintf("10.%d.%d.0/24", secondOctet, thirdOctet)

	return subnet, nil
}

// ReloadServerConfig executes the command specified in SERVER_RELOAD_CMD environment variable
func ReloadServerConfig() error {
	cmdString := os.Getenv("SERVER_RELOAD_CMD")
	if cmdString == "" {
		// No reload command specified, nothing to do so lets just return without error
		return nil
	}

	log.Infof("Executing server reload command: %s", cmdString)

	// Execute the command through shell to preserve the exact format
	cmd := exec.Command("bash", "-c", cmdString)

	// Capture output for logging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to execute server reload command: %w, output: %s", err, string(output))
	}

	log.Infof("Server reload command executed successfully")
	if len(output) > 0 {
		log.Debugf("Command output: %s", string(output))
	}

	return nil
}
