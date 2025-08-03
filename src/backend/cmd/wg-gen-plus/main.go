//
// Copyright 2025 Lindsay Harvey - https://github.com/vk2amv
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wg-gen-plus/api"
	"wg-gen-plus/auth"
	"wg-gen-plus/core"
	"wg-gen-plus/storage"
	"wg-gen-plus/util"
	"wg-gen-plus/version"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const help = `Wg Gen Plus is a comprehensive web based configuration generator for WireGuard
Usage: wg-gen-plus [options]

Helpers:
  --version       display the version number of Wg Gen Plus
  --help          display this help message

Options:
  --config=<file>            file name of optional Wg Gen Plus configuration file
  --db-dir=<path>            directory for the database file (default: /var/lib/wg-gen-plus)
  --wg-interface=<name>      name of the WireGuard interface (default: wg0)
  --wg-conf-path=<name>      path to the WireGuard configuration directory (default: /etc/wireguard/)
  --port=<port>              port to run the web server on (default: 8080)
  --server=<address>         address to bind the web server to (default: 0.0.0.0)
  --use-defaults=true|false  use all default values for server (default: false - enable for testing only)
`

// Default configuration values
const (
	// WireGuard configuration
	DefaultWgConfPath  = "/etc/wireguard"
	DefaultWgInterface = "wg0"

	// Application paths
	DefaultWgGenConf = "/etc/wg-gen-plus"
	DefaultDbFileDir = "/var/lib/wg-gen-plus"

	// Server settings
	DefaultServerAddress = "0.0.0.0"
	DefaultServerPort    = "80"

	// Application settings
	DefaultGinMode  = "release"
	DefaultAuthType = "local"
)

var WgConfigFile string

var (
	cacheDb = cache.New(60*time.Minute, 10*time.Minute)
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
}

func main() {

	// Check if --help or -h is in the raw args
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Print(help)
			os.Exit(0)
		}
		if arg == "--version" || arg == "-v" {
			fmt.Println(version.Version)
			os.Exit(0)
		}
	}

	log.Infof("Starting Wg Gen Plus version: %s", version.Version)

	// Vars for config flags
	var (
		configPath      string
		configPathflag  string
		configFile      string
		configFileflag  string
		dbDirFlag       string
		wgInterface     string
		wgInterfaceflag string
		wgConfPathflag  string
		portflag        string
		serverflag      string
		ginmodeflag     bool
		useDefaults     bool
		err             error
	)

	flag.StringVar(&configPathflag, "configdir", "", "Path to the config file directory")
	flag.StringVar(&configFileflag, "config", "", "Wg Gen Plus config file to use")
	flag.StringVar(&configFileflag, "C", "", "Shorthand for --config")
	flag.StringVar(&dbDirFlag, "db-dir", "", "Directory for the database file")
	flag.StringVar(&wgInterfaceflag, "wg-interface", "", "Name of the WireGuard interface")
	flag.StringVar(&wgConfPathflag, "wg-conf-path", "", "Path to the WireGuard configuration directory")
	flag.StringVar(&portflag, "port", "", "Port to run the web server on")
	flag.StringVar(&serverflag, "server", "", "Address to bind the web server to")
	flag.BoolVar(&ginmodeflag, "debug", false, "Set Gin mode to debug")
	flag.BoolVar(&useDefaults, "use-defaults", false, "Use all default values for server configuration (For testing only)")
	flag.Parse()

	if !useDefaults {

		// If configPathflag is set, use it, otherwise use the environment variable
		configPath = os.Getenv("WG_GEN_CONF_DIR")
		if configPathflag != "" {
			configPath = configPathflag
		} else if configPath == "" && configFileflag == "" && os.Getenv("WG_GEN_CONF_FILE") == "" {
			// If no config path variable exists AND no config file flag AND no env variable,
			// fallback to default config file location and name
			configPath = DefaultWgGenConf
			wgInterface = DefaultWgInterface
			os.Setenv("WG_INTERFACE_NAME", DefaultWgInterface)
			os.Setenv("WG_GEN_CONF_DIR", configPath)

			configFile = filepath.Join(configPath, fmt.Sprintf("wg-gen-plus-%s.conf", wgInterface))
			os.Setenv("WG_GEN_CONF_FILE", configFile)

			// Don't try to load the default config file if it doesn't exist
			if util.FileExists(configFile) {
				err := godotenv.Load(configFile)
				if err != nil {
					log.WithFields(log.Fields{
						"err":  err,
						"file": configFile,
					}).Warning("Default config file exists but could not be loaded")
				} else {
					log.WithField("file", configFile).Info("Loaded default config file")
				}
			} else {
				log.WithField("file", configFile).Fatal("No default config file found, spitting dummy out of crib")
			}
		}
	}

	// Override any config values from flags
	if configFileflag != "" {
		// Load the environment file directly if provided via --config
		err := godotenv.Load(configFileflag)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"file": configFileflag,
			}).Fatal("failed to load config file")
		}
		log.WithField("file", configFileflag).Info("Loaded config file")

		os.Setenv("WG_GEN_CONF_FILE", configFileflag)
	}
	if dbDirFlag != "" {
		os.Setenv("DB_FILE_DIR", dbDirFlag)
	}
	if wgInterfaceflag != "" {
		os.Setenv("WG_INTERFACE_NAME", wgInterfaceflag)
	}
	if wgConfPathflag != "" {
		os.Setenv("WG_CONF_DIR", wgConfPathflag)
	}
	if portflag != "" {
		os.Setenv("PORT", portflag)
	}
	if serverflag != "" {
		os.Setenv("SERVER", serverflag)
	}
	if ginmodeflag {
		os.Setenv("GIN_MODE", "debug")
	}

	// Set default environment values if USE_DEFAULTS is true
	if useDefaults {
		log.Info("Using default configuration values")
		configPath = DefaultWgGenConf
		setDefaultsIfRequested()
	}

	// Set file path variables from environment
	wgInterface = os.Getenv("WG_INTERFACE_NAME")
	dbDir := os.Getenv("DB_FILE_DIR")
	wgConfPath := os.Getenv("WG_CONF_DIR")
	configFile = os.Getenv("WG_GEN_CONF_FILE")

	// Assemble file names for Wg Gen Plus config
	if configFile == "" {
		configFile = filepath.Join(configPath, fmt.Sprintf("wg-gen-plus-%s.conf", wgInterface))
	}

	// Load additional environment variables from config file if it exists, eg SMTP settings
	if !useDefaults {
		envMap, err := godotenv.Read(configFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.WithField("path", configFile).Info("config file for extended variables not found")
			} else {
				log.WithFields(log.Fields{
					"err":  err,
					"path": configFile,
				}).Fatal("failed to load config file")
			}
		} else {
			// Only set variables that are NOT already set previously so we don't blast any overrides
			for key, value := range envMap {
				if _, exists := os.LookupEnv(key); !exists {
					os.Setenv(key, value)
				}
			}
		}
	}

	// Set other filename paths
	WgConfigFile = filepath.Join(wgConfPath, fmt.Sprintf("%s.conf", wgInterface))
	core.WgConfigFile = WgConfigFile
	dbFile := filepath.Join(dbDir, fmt.Sprintf("wg-gen-plus-%s.db", wgInterface))

	// Log file paths to console
	log.WithFields(log.Fields{
		"configFile":   configFile,
		"dbFile":       dbFile,
		"WgConfigFile": WgConfigFile,
	}).Info("Using the following file paths")

	if os.Getenv("WG_STATS_API") != "" {
		log.WithField("WG_STATS_API", os.Getenv("WG_STATS_API")).Info("WG_STATS_API environment variable found")
	}

	// Validate required environment variables exist before writing stuff to disk, spit dummy if any are missing here
	missing := []string{}
	if configFile == "" {
		missing = append(missing, "configFile")
	}
	if dbDir == "" {
		missing = append(missing, "dbDir")
	}
	if wgConfPath == "" {
		missing = append(missing, "wgConfPath")
	}
	if len(missing) > 0 {
		log.WithFields(log.Fields{
			"missing": missing,
		}).Fatal("required configuration value(s) are missing")
	}

	// Check if the WireGuard config directory exists
	if !util.DirectoryExists(wgConfPath) {
		log.WithFields(log.Fields{
			"dir": wgConfPath,
		}).Fatalf("WireGuard config directory does not exist: %s. Please create it or set WG_CONF_DIR to a valid directory.", wgConfPath)
	}

	// Check if the DB directory exists
	if !util.DirectoryExists(dbDir) {
		log.WithFields(log.Fields{
			"dir": dbDir,
		}).Fatalf("Database directory does not exist: %s. Please create it or set DB_FILE_DIR to a valid directory.", dbDir)
	}

	// Initialize database
	err = storage.InitStorage(dbFile)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"db_file": dbFile,
		}).Fatal("failed to initialize SQLite storage")
	}

	if os.Getenv("GIN_MODE") == "debug" {
		// set gin release debug
		gin.SetMode(gin.DebugMode)
	} else {
		// set gin release mode
		gin.SetMode(gin.ReleaseMode)
		// disable console color
		gin.DisableConsoleColor()
		// log level info
		log.SetLevel(log.InfoLevel)
	}

	// dump wg config file
	err = core.UpdateServerConfigWg()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to dump wg config file")
	}

	// creates a gin router with default middleware: logger and recovery (crash-free) middleware
	app := gin.Default()

	// cors middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization", util.AuthTokenHeaderName)
	app.Use(cors.New(config))

	// protection middleware
	app.Use(helmet.Default())

	// add cache storage to gin app
	app.Use(func(ctx *gin.Context) {
		ctx.Set("cache", cacheDb)
		ctx.Next()
	})

	// serve static files
	app.Use(static.Serve("/", static.LocalFile("./ui/dist", false)))

	// setup Oauth2 client only if not using local auth
	if !auth.IsLocalAuth() {
		log.Info("Setting up OAuth2 client")
		oauth2Client, err := auth.GetAuthProvider()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("failed to setup Oauth2")
		}

		app.Use(func(ctx *gin.Context) {
			ctx.Set("oauth2Client", oauth2Client)
			ctx.Next()
		})
	} else {
		log.Info("Using local authentication, skipping OAuth2 setup")
		// For API endpoints that expect an oauth2Client, we'll provide a dummy
		app.Use(func(ctx *gin.Context) {
			// Setting nil is fine since we check IsLocalAuth() before using it
			ctx.Set("oauth2Client", nil)
			ctx.Next()
		})
	}

	// apply api routes public
	api.ApplyRoutes(app, false)

	// simple middleware to check auth
	app.Use(func(c *gin.Context) {
		// Skip auth for login and other public endpoints
		if strings.Contains(c.Request.URL.Path, "/api/v1.0/auth/") {
			c.Next()
			return
		}

		cacheDb := c.MustGet("cache").(*cache.Cache)
		token := c.Request.Header.Get(util.AuthTokenHeaderName)

		// For local auth, the token maps to a user ID
		if auth.IsLocalAuth() {
			// If no token, reject the request
			if token == "" {
				// avoid 401 page for refresh after logout or for static assets
				if !strings.Contains(c.Request.URL.Path, "/api/") {
					c.Redirect(301, "/index.html")
					return
				}
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// Check if token exists in cache
			userID, found := cacheDb.Get(token)
			if !found {
				// Token not found or expired
				if !strings.Contains(c.Request.URL.Path, "/api/") {
					c.Redirect(301, "/index.html")
					return
				}
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// Set user ID in context for use in handlers
			userIDStr, ok := userID.(string)
			if !ok {
				log.Error("User ID in cache is not a string")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			log.Debugf("User authenticated with ID: %s", userIDStr)
			c.Set("userID", userIDStr)
			c.Next()
			return
		}

		// OAuth2 authentication logic
		oauth2Token, exists := cacheDb.Get(token)
		if exists && oauth2Token.(*oauth2.Token).AccessToken == token {
			// will be accessible in auth endpoints
			c.Set("oauth2Token", oauth2Token)
			c.Next()
			return
		}

		// avoid 401 page for refresh after logout
		if !strings.Contains(c.Request.URL.Path, "/api/") {
			c.Redirect(301, "/index.html")
			return
		}

		c.AbortWithStatus(http.StatusUnauthorized)
	})

	// apply api router private
	api.ApplyRoutes(app, true)

	// NoRoute handler for SPA routing - ensures all frontend routes work
	app.NoRoute(func(c *gin.Context) {
		// If the request is for an API endpoint, return 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// For all other routes, serve the index.html
		// This enables client-side routing
		c.File("./ui/dist/index.html") // Using your static file path
	})

	err = app.Run(fmt.Sprintf("%s:%s", os.Getenv("SERVER"), os.Getenv("PORT")))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to start server")
	}
}

func setDefaultsIfRequested() {
	os.Setenv("WG_CONF_DIR", DefaultWgConfPath)
	os.Setenv("WG_INTERFACE_NAME", DefaultWgInterface)
	os.Setenv("WG_GEN_CONF_DIR", DefaultWgGenConf)
	os.Setenv("DB_FILE_DIR", DefaultDbFileDir)
	os.Setenv("SERVER", DefaultServerAddress)
	os.Setenv("PORT", DefaultServerPort)
	os.Setenv("GIN_MODE", DefaultGinMode)
	os.Setenv("AUTH_TYPE", DefaultAuthType)
}
