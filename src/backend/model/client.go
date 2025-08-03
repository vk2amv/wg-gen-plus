package model

import (
	"fmt"
	"time"
	"wg-gen-plus/util"
)

// Client structure
type Client struct {
	Id                              string    `json:"id"`
	Name                            string    `json:"name"`
	Email                           string    `json:"email"`
	Enable                          bool      `json:"enable"`
	Site2Site                       bool      `json:"site2site"`
	IgnorePersistentKeepalive       bool      `json:"ignorePersistentKeepalive"`
	KeepaliveDisabled               bool      `json:"keepaliveDisabled"`
	KeepaliveInterval               int       `json:"keepaliveInterval"`
	UseRemoteDNS                    bool      `json:"useRemoteDNS"`
	Site2SiteEndpointOptionsEnabled bool      `json:"site2siteEndpointOptionsEnabled"`
	Site2SiteEndpoint               string    `json:"site2SiteEndpoint"`
	Site2SiteEndpointPort           int       `json:"site2SiteEndpointPort"`
	Site2SiteEndpointListenPort     int       `json:"site2SiteEndpointListenPort"`
	LANIPs                          []string  `json:"lanIPs"`
	Table                           string    `json:"table"`
	PresharedKey                    string    `json:"presharedKey"`
	AllowedIPs                      []string  `json:"allowedIPs"`
	Address                         []string  `json:"address"`
	Tags                            []string  `json:"tags"`
	PrivateKey                      string    `json:"privateKey"`
	PublicKey                       string    `json:"publicKey"`
	CreatedBy                       string    `json:"createdBy"`
	UpdatedBy                       string    `json:"updatedBy"`
	Created                         time.Time `json:"created"`
	Updated                         time.Time `json:"updated"`
}

// IsValid check if model is valid
func (a Client) IsValid() []error {
	errs := make([]error, 0)

	// check if the name empty
	if a.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	// check the name field is between 3 to 40 chars
	if len(a.Name) < 2 || len(a.Name) > 40 {
		errs = append(errs, fmt.Errorf("name field must be between 2-40 chars"))
	}
	// email is not required, but if provided must match regex
	if a.Email != "" {
		if !util.RegexpEmail.MatchString(a.Email) {
			errs = append(errs, fmt.Errorf("email %s is invalid", a.Email))
		}
	}

	// If Site2Site is true, then LANIPs are required
	if a.Site2Site {
		// Check if LANIPs is empty when Site2Site is true
		if len(a.LANIPs) == 0 {
			errs = append(errs, fmt.Errorf("LANIPs are required when Site2Site is enabled"))
		}
	}

	// Site2SiteEndpoint/Port validation logic
	if a.Site2SiteEndpoint == "" {
		if a.Site2SiteEndpointPort != 0 {
			errs = append(errs, fmt.Errorf("Site2SiteEndpointPort must be unset when Site2SiteEndpoint is empty"))
		}
		if a.Site2SiteEndpointListenPort != 0 {
			errs = append(errs, fmt.Errorf("Site2SiteEndpointListenPort must be unset when Site2SiteEndpoint is empty"))
		}
	} else {
		// Site2SiteEndpoint is NOT empty
		if a.Site2SiteEndpointListenPort <= 0 || a.Site2SiteEndpointListenPort > 65535 {
			errs = append(errs, fmt.Errorf("Site2SiteEndpointListenPort %d is invalid", a.Site2SiteEndpointListenPort))
		}
		if a.Site2SiteEndpointPort != 0 {
			if a.Site2SiteEndpointPort <= 0 || a.Site2SiteEndpointPort > 65535 {
				errs = append(errs, fmt.Errorf("Site2SiteEndpointPort %d is invalid", a.Site2SiteEndpointPort))
			}
			// If Port is set, ListenPort must also be valid
			if a.Site2SiteEndpointListenPort <= 0 || a.Site2SiteEndpointListenPort > 65535 {
				errs = append(errs, fmt.Errorf("Site2SiteEndpointListenPort %d is invalid (required when Site2SiteEndpointPort is set)", a.Site2SiteEndpointListenPort))
			}
		}
	}

	// If IgnorePersistentKeepalive is true and KeepaliveDisabled is false,
	// then KeepaliveInterval must be set and a positive integer
	if a.IgnorePersistentKeepalive && !a.KeepaliveDisabled {
		if a.KeepaliveInterval <= 0 {
			errs = append(errs, fmt.Errorf("KeepaliveInterval must be set to a positive integer when IgnorePersistentKeepalive is true and KeepaliveDisabled is false"))
		}
	}

	// check if the lanIPs are valid (required if Site2Site is true, optional otherwise)
	for _, lanIP := range a.LANIPs {
		if !util.IsValidCidr(lanIP) {
			errs = append(errs, fmt.Errorf("lanIP %s is invalid", lanIP))
		}
	}

	// check if the allowedIPs empty
	if len(a.AllowedIPs) == 0 {
		errs = append(errs, fmt.Errorf("allowedIPs field is required"))
	}
	// check if the allowedIPs are valid
	for _, allowedIP := range a.AllowedIPs {
		if !util.IsValidCidr(allowedIP) {
			errs = append(errs, fmt.Errorf("allowedIP %s is invalid", allowedIP))
		}
	}
	// check if the address empty
	if len(a.Address) == 0 {
		errs = append(errs, fmt.Errorf("address field is required"))
	}
	// check if the address are valid
	for _, address := range a.Address {
		if !util.IsValidCidr(address) {
			errs = append(errs, fmt.Errorf("address %s is invalid", address))
		}
	}

	return errs
}
