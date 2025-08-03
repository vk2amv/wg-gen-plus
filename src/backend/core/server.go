package core

import (
	"errors"
	"net"
	"os"
	"time"
	"wg-gen-plus/model"
	"wg-gen-plus/storage"
	"wg-gen-plus/template"
	"wg-gen-plus/util"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var WgConfigFile string

// ReadServer object, create default one
func ReadServer() (*model.Server, error) {
	server, err := storage.LoadServer()
	if err == nil && server != nil {
		return server, nil
	}

	// If not found, create default
	server = &model.Server{}
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	// Generate a random IPv6 ULA subnet
	ipv6Subnet, err := util.GenerateRandomULASubnet()
	if err != nil {
		return nil, err
	}
	// Get the first usable IPv6 address from the generated subnet for the server
	ipv6, _, err := net.ParseCIDR(ipv6Subnet)
	if err != nil {
		return nil, err
	}
	// Increment the IP to get the first usable address in the subnet
	util.Inc(ipv6)
	ipv6ServerAddress := ipv6.String() + "/64"

	// Generate a random IPv4 subnet
	ipv4Subnet, err := util.GenerateRandomIPv4Subnet()
	if err != nil {
		return nil, err
	}
	// Get the first usable IPv4 address from the generated subnet
	ipv4, _, err := net.ParseCIDR(ipv4Subnet)
	if err != nil {
		return nil, err
	}
	// Increment the IP to get the first usable address in the subnet
	util.Inc(ipv4)
	ipv4ServerAddress := ipv4.String() + "/24"

	server.PrivateKey = key.String()
	server.PublicKey = key.PublicKey().String()
	server.Endpoint = "wireguard.example.com:123"
	server.ListenPort = 51820
	server.Address = []string{ipv6ServerAddress, ipv4ServerAddress}
	server.Dns = []string{ipv6.String(), ipv4.String()}
	server.AllowedIPs = []string{ipv6Subnet, ipv4Subnet}
	server.PersistentKeepalive = 25
	server.Mtu = 0
	server.Created = time.Now().UTC()
	server.Updated = server.Created

	err = storage.SaveServer(server)
	if err != nil {
		return nil, err
	}
	err = UpdateServerConfigWg()
	if err != nil {
		return nil, err
	}
	return server, nil
}

// UpdateServer keep private values from existing one
func UpdateServer(server *model.Server) (*model.Server, error) {
	current, err := storage.LoadServer()
	if err != nil {
		return nil, err
	}
	errs := server.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("server validation error")
		}
		return nil, errors.New("failed to validate server")
	}
	server.PrivateKey = current.PrivateKey
	server.PublicKey = current.PublicKey
	server.Updated = time.Now().UTC()

	err = storage.SaveServer(server)
	if err != nil {
		return nil, err
	}
	server, err = storage.LoadServer()
	if err != nil {
		return nil, err
	}
	return server, UpdateServerConfigWg()
}

// UpdateServerConfigWg in wg format
func UpdateServerConfigWg() error {
	// Check if WgConfigFile is empty
	if WgConfigFile == "" {
		return errors.New("WireGuard config file path is empty")
	}

	clients, err := ReadClients()
	if err != nil {
		return err
	}

	server, err := ReadServer()
	if err != nil {
		return err
	}

	// Get environment variables for hook scripts
	preUpHook := os.Getenv("SERVER_PREUP_HOOK")
	postUpHook := os.Getenv("SERVER_POSTUP_HOOK")
	preDownHook := os.Getenv("SERVER_PREDOWN_HOOK")
	postDownHook := os.Getenv("SERVER_POSTDOWN_HOOK")

	// Use the global WgConfigFile variable
	_, err = template.DumpServerWg(clients, server, preUpHook, postUpHook, preDownHook, postDownHook, WgConfigFile)
	if err != nil {
		return err
	}

	return nil
}

// GetAllReservedIps the list of all reserved IPs, client and server
func GetAllReservedIps() ([]string, error) {
	clients, err := ReadClients()
	if err != nil {
		return nil, err
	}

	server, err := ReadServer()
	if err != nil {
		return nil, err
	}

	reserverIps := make([]string, 0)

	for _, client := range clients {
		for _, cidr := range client.Address {
			ip, err := util.GetIpFromCidr(cidr)
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"cidr": cidr,
				}).Error("failed to get IP from CIDR")
			} else {
				reserverIps = append(reserverIps, ip)
			}
		}
	}

	for _, cidr := range server.Address {
		ip, err := util.GetIpFromCidr(cidr)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"cidr": cidr,
			}).Error("failed to get IP from CIDR")
		} else {
			reserverIps = append(reserverIps, ip)
		}
	}

	return reserverIps, nil
}

// ReadWgConfigFile return content of wireguard config file
func ReadWgConfigFile() ([]byte, error) {
	return util.ReadFile(WgConfigFile)
}
