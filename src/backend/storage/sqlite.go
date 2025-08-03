package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"wg-gen-plus/model"

	uuid "github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitStorage(dbFile string) error {
	fmt.Println("Opening SQLite DB at:", dbFile)
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	// Create schema
	err = createSchema()
	if err != nil {
		return err
	}

	// Create default admin user if no users exist
	err = createDefaultAdminIfNeeded()
	if err != nil {
		return err
	}

	return nil
}

func createSchema() error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS clients (
		id TEXT PRIMARY KEY,
		name TEXT,
		email TEXT,
		enable INTEGER,
		site2site INTEGER,
		ignore_persistent_keepalive INTEGER,
		keepalive_disabled INTEGER,
		keepalive_interval INTEGER,
		use_remote_dns INTEGER,
		site2site_endpoint_options_enabled INTEGER,
		site2site_endpoint TEXT,
		site2site_endpoint_port INTEGER,
		site2site_endpoint_listen_port INTEGER,
		lan_ips TEXT,
		table_name TEXT,
		preshared_key TEXT,
		allowed_ips TEXT,
		address TEXT,
		tags TEXT,
		private_key TEXT,
		public_key TEXT,
		created_by TEXT,
		updated_by TEXT,
		created TEXT,
		updated TEXT
	);
	CREATE TABLE IF NOT EXISTS server (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		address TEXT,
		listen_port INTEGER,
		mtu INTEGER,
		private_key TEXT,
		public_key TEXT,
		endpoint TEXT,
		persistent_keepalive INTEGER,
		dns TEXT,
		allowed_ips TEXT,
		table_name TEXT,
		updated_by TEXT,
		created TEXT,
		updated TEXT
	);
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		email TEXT,
		password TEXT,
    	is_admin INTEGER
	);
	`)
	return err
}

// SaveClient saves a client to the database
func SaveClient(c *model.Client) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	// Encode slices as JSON
	allowedIPsJSON, _ := json.Marshal(c.AllowedIPs)
	addressJSON, _ := json.Marshal(c.Address)
	tagsJSON, _ := json.Marshal(c.Tags)
	lanIPsJSON, _ := json.Marshal(c.LANIPs)

	_, err := db.Exec(`
    INSERT INTO clients (
        id, name, email, enable, site2site, ignore_persistent_keepalive, 
        keepalive_disabled, keepalive_interval, use_remote_dns,
        site2site_endpoint_options_enabled,
        site2site_endpoint, site2site_endpoint_port, site2site_endpoint_listen_port,
        lan_ips, table_name, preshared_key, allowed_ips, address, tags,
        private_key, public_key, created_by, updated_by, created, updated
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
        name=excluded.name,
        email=excluded.email,
        enable=excluded.enable,
        site2site=excluded.site2site,
        ignore_persistent_keepalive=excluded.ignore_persistent_keepalive,
        keepalive_disabled=excluded.keepalive_disabled,
        keepalive_interval=excluded.keepalive_interval,
        use_remote_dns=excluded.use_remote_dns,
        site2site_endpoint_options_enabled=excluded.site2site_endpoint_options_enabled,
        site2site_endpoint=excluded.site2site_endpoint,
        site2site_endpoint_port=excluded.site2site_endpoint_port,
        site2site_endpoint_listen_port=excluded.site2site_endpoint_listen_port,
        lan_ips=excluded.lan_ips,
        table_name=excluded.table_name,
        preshared_key=excluded.preshared_key,
        allowed_ips=excluded.allowed_ips,
        address=excluded.address,
        tags=excluded.tags,
        private_key=excluded.private_key,
        public_key=excluded.public_key,
        created_by=excluded.created_by,
        updated_by=excluded.updated_by,
        created=excluded.created,
        updated=excluded.updated
`, c.Id, c.Name, c.Email, boolToInt(c.Enable), boolToInt(c.Site2Site),
		boolToInt(c.IgnorePersistentKeepalive), boolToInt(c.KeepaliveDisabled), c.KeepaliveInterval,
		boolToInt(c.UseRemoteDNS), boolToInt(c.Site2SiteEndpointOptionsEnabled),
		c.Site2SiteEndpoint, c.Site2SiteEndpointPort, c.Site2SiteEndpointListenPort,
		string(lanIPsJSON),
		c.Table, c.PresharedKey, string(allowedIPsJSON),
		string(addressJSON), string(tagsJSON),
		c.PrivateKey, c.PublicKey, c.CreatedBy, c.UpdatedBy,
		c.Created.Format(time.RFC3339), c.Updated.Format(time.RFC3339))

	return err
}

// LoadClient loads a client by id
func LoadClient(id string) (*model.Client, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	row := db.QueryRow(`SELECT 
        id, name, email, enable, site2site, ignore_persistent_keepalive, 
        keepalive_disabled, keepalive_interval, use_remote_dns,
        site2site_endpoint_options_enabled,
        site2site_endpoint, site2site_endpoint_port, site2site_endpoint_listen_port,
        lan_ips, table_name, preshared_key, allowed_ips, address, tags,
        private_key, public_key, created_by, updated_by, created, updated
    FROM clients WHERE id = ?`, id)

	var c model.Client
	var allowedIPsJSON, addressJSON, tagsJSON, lanIPsJSON string
	var createdStr, updatedStr string
	var enableInt, site2siteInt, ignorePKInt, keepaliveDisabledInt, useRemoteDNSInt, endpointOptionsEnabledInt int

	err := row.Scan(
		&c.Id, &c.Name, &c.Email, &enableInt, &site2siteInt, &ignorePKInt,
		&keepaliveDisabledInt, &c.KeepaliveInterval, &useRemoteDNSInt,
		&endpointOptionsEnabledInt,
		&c.Site2SiteEndpoint, &c.Site2SiteEndpointPort, &c.Site2SiteEndpointListenPort,
		&lanIPsJSON, &c.Table, &c.PresharedKey, &allowedIPsJSON, &addressJSON, &tagsJSON,
		&c.PrivateKey, &c.PublicKey, &c.CreatedBy, &c.UpdatedBy, &createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}

	// Convert integers to booleans
	c.Enable = enableInt != 0
	c.Site2Site = site2siteInt != 0
	c.IgnorePersistentKeepalive = ignorePKInt != 0
	c.KeepaliveDisabled = keepaliveDisabledInt != 0
	c.UseRemoteDNS = useRemoteDNSInt != 0
	c.Site2SiteEndpointOptionsEnabled = endpointOptionsEnabledInt != 0

	// Unmarshal JSON strings to slices
	_ = json.Unmarshal([]byte(allowedIPsJSON), &c.AllowedIPs)
	_ = json.Unmarshal([]byte(addressJSON), &c.Address)
	_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
	_ = json.Unmarshal([]byte(lanIPsJSON), &c.LANIPs)

	// Parse timestamps
	c.Created, _ = time.Parse(time.RFC3339, createdStr)
	c.Updated, _ = time.Parse(time.RFC3339, updatedStr)

	return &c, nil
}

// DeleteClient deletes a client by id
func DeleteClient(id string) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	_, err := db.Exec("DELETE FROM clients WHERE id = ?", id)
	return err
}

// LoadAllClients loads all clients from the database
func LoadAllClients() ([]*model.Client, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	rows, err := db.Query(`SELECT 
        id, name, email, enable, site2site, ignore_persistent_keepalive, 
        keepalive_disabled, keepalive_interval, use_remote_dns,
        site2site_endpoint_options_enabled,
        site2site_endpoint, site2site_endpoint_port, site2site_endpoint_listen_port,
        lan_ips, table_name, preshared_key, allowed_ips, address, tags,
        private_key, public_key, created_by, updated_by, created, updated
    FROM clients`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []*model.Client{}
	for rows.Next() {
		var c model.Client
		var allowedIPsJSON, addressJSON, tagsJSON, lanIPsJSON string
		var createdStr, updatedStr string
		var enableInt, site2siteInt, ignorePKInt, keepaliveDisabledInt, useRemoteDNSInt, endpointOptionsEnabledInt int

		err := rows.Scan(
			&c.Id, &c.Name, &c.Email, &enableInt, &site2siteInt, &ignorePKInt,
			&keepaliveDisabledInt, &c.KeepaliveInterval, &useRemoteDNSInt,
			&endpointOptionsEnabledInt,
			&c.Site2SiteEndpoint, &c.Site2SiteEndpointPort, &c.Site2SiteEndpointListenPort,
			&lanIPsJSON, &c.Table, &c.PresharedKey, &allowedIPsJSON, &addressJSON, &tagsJSON,
			&c.PrivateKey, &c.PublicKey, &c.CreatedBy, &c.UpdatedBy, &createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}

		// Convert integers to booleans
		c.Enable = enableInt != 0
		c.Site2Site = site2siteInt != 0
		c.IgnorePersistentKeepalive = ignorePKInt != 0
		c.KeepaliveDisabled = keepaliveDisabledInt != 0
		c.UseRemoteDNS = useRemoteDNSInt != 0
		c.Site2SiteEndpointOptionsEnabled = endpointOptionsEnabledInt != 0

		// Unmarshal JSON strings to slices
		_ = json.Unmarshal([]byte(allowedIPsJSON), &c.AllowedIPs)
		_ = json.Unmarshal([]byte(addressJSON), &c.Address)
		_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
		_ = json.Unmarshal([]byte(lanIPsJSON), &c.LANIPs)

		// Parse timestamps
		c.Created, _ = time.Parse(time.RFC3339, createdStr)
		c.Updated, _ = time.Parse(time.RFC3339, updatedStr)

		clients = append(clients, &c)
	}

	return clients, nil
}

// SaveServer saves the server config to the database
func SaveServer(s *model.Server) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	addressJSON, _ := json.Marshal(s.Address)
	dnsJSON, _ := json.Marshal(s.Dns)
	allowedIPsJSON, _ := json.Marshal(s.AllowedIPs)

	_, err := db.Exec(`
    INSERT INTO server (
        id, address, listen_port, mtu, private_key, public_key, endpoint,
        persistent_keepalive, dns, allowed_ips, updated_by, created, updated
    ) VALUES (
        1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
    )
    ON CONFLICT(id) DO UPDATE SET
        address=excluded.address,
        listen_port=excluded.listen_port,
        mtu=excluded.mtu,
        private_key=excluded.private_key,
        public_key=excluded.public_key,
        endpoint=excluded.endpoint,
        persistent_keepalive=excluded.persistent_keepalive,
        dns=excluded.dns,
        allowed_ips=excluded.allowed_ips,
        updated_by=excluded.updated_by,
        created=excluded.created,
        updated=excluded.updated
    `, string(addressJSON), s.ListenPort, s.Mtu, s.PrivateKey, s.PublicKey, s.Endpoint,
		s.PersistentKeepalive, string(dnsJSON), string(allowedIPsJSON),
		s.UpdatedBy,
		s.Created.Format(time.RFC3339), s.Updated.Format(time.RFC3339))
	return err
}

// LoadServer loads the server config from the database
func LoadServer() (*model.Server, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	row := db.QueryRow(`SELECT
        address, listen_port, mtu, private_key, public_key, endpoint,
        persistent_keepalive, dns, allowed_ips, updated_by, created, updated
        FROM server WHERE id = 1`)
	var s model.Server
	var addressJSON, dnsJSON, allowedIPsJSON string
	var createdStr, updatedStr string

	err := row.Scan(
		&addressJSON, &s.ListenPort, &s.Mtu, &s.PrivateKey, &s.PublicKey, &s.Endpoint,
		&s.PersistentKeepalive, &dnsJSON, &allowedIPsJSON, &s.UpdatedBy,
		&createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(addressJSON), &s.Address)
	_ = json.Unmarshal([]byte(dnsJSON), &s.Dns)
	_ = json.Unmarshal([]byte(allowedIPsJSON), &s.AllowedIPs)
	s.Created, _ = time.Parse(time.RFC3339, createdStr)
	s.Updated, _ = time.Parse(time.RFC3339, updatedStr)
	return &s, nil
}

// SaveUser creates or updates a user in the database
func SaveUser(user *model.User) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	_, err := db.Exec(`
        INSERT INTO users (id, name, email, password, is_admin)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            name=excluded.name,
            email=excluded.email,
            password=excluded.password,
            is_admin=excluded.is_admin
    `, user.Sub, user.Name, user.Email, user.Password, boolToInt(user.IsAdmin))

	return err
}

// LoadUser retrieves a user by their ID
func LoadUser(id string) (*model.User, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	row := db.QueryRow(`
        SELECT id, name, email, password, is_admin
        FROM users
        WHERE id = ?
    `, id)

	var user model.User
	var isAdminInt int
	err := row.Scan(&user.Sub, &user.Name, &user.Email, &user.Password, &isAdminInt)
	if err != nil {
		return nil, err
	}

	user.IsAdmin = isAdminInt != 0

	return &user, nil
}

// LoadAllUsers retrieves all users from the database
func LoadAllUsers() ([]*model.User, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	rows, err := db.Query(`
        SELECT id, name, email, password, is_admin
        FROM users
        ORDER BY name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		var isAdminInt int
		err := rows.Scan(&user.Sub, &user.Name, &user.Email, &user.Password, &isAdminInt)
		if err != nil {
			return nil, err
		}
		user.IsAdmin = isAdminInt != 0
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser removes a user from the database
func DeleteUser(id string) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// createDefaultAdminIfNeeded creates a default admin user if no users exist
func createDefaultAdminIfNeeded() error {
	if db == nil {
		return errors.New("database not initialized")
	}

	// Check if any users exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	// If no users exist, create a default admin
	if count == 0 {
		fmt.Println("No users found. Creating default admin user...")

		// Generate a random password
		password := generateRandomPassword(12)
		fmt.Println("Generated admin password:", password)

		// Hash password using bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}

		// Generate UUID for admin user
		u, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("failed to generate UUID: %v", err)
		}

		// Create admin user
		adminUser := &model.User{
			Sub:      u.String(),
			Name:     "admin",
			Email:    "admin@example.com",
			Password: string(hashedPassword),
			IsAdmin:  true,
		}

		err = SaveUser(adminUser)
		if err != nil {
			return fmt.Errorf("failed to save admin user: %v", err)
		}

		fmt.Println("Default admin user created:")
		fmt.Println("  Username: admin")
		fmt.Println("  Password:", password)
		fmt.Println("  Please change this password after logging in!")
	}

	return nil
}

// Helper function to generate a random password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator once

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
