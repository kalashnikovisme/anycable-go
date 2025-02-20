package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
)

func (c *Config) Presets() []string {
	if c.UserPresets != nil {
		return c.UserPresets
	}

	return detectPresetsFromEnv()
}

func (c *Config) LoadPresets() error {
	presets := c.Presets()

	if len(presets) == 0 {
		return nil
	}

	log.WithField("context", "config").Infof("Load presets: %s", strings.Join(presets, ","))

	defaults := NewConfig()

	for _, preset := range presets {
		switch preset {
		case "fly":
			if err := c.loadFlyPreset(&defaults); err != nil {
				return err
			}
		case "heroku":
			if err := c.loadHerokuPreset(&defaults); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Config) loadFlyPreset(defaults *Config) error {
	if c.Host == defaults.Host {
		c.Host = "0.0.0.0"
	}

	region, ok := os.LookupEnv("FLY_REGION")

	if !ok {
		return errors.New("FLY_REGION env is missing")
	}

	appName, ok := os.LookupEnv("FLY_APP_NAME")

	if !ok {
		return errors.New("FLY_APP_NAME env is missing")
	}

	if c.EmbeddedNats.ServiceAddr == defaults.EmbeddedNats.ServiceAddr {
		c.EmbeddedNats.ServiceAddr = "nats://0.0.0.0:4222"
	}

	if c.EmbeddedNats.ClusterAddr == defaults.EmbeddedNats.ClusterAddr {
		c.EmbeddedNats.ClusterAddr = "nats://0.0.0.0:5222"
	}

	if c.EmbeddedNats.ClusterName == defaults.EmbeddedNats.ClusterName {
		c.EmbeddedNats.ClusterName = fmt.Sprintf("%s-%s-cluster", appName, region)
	}

	if c.EmbeddedNats.Routes == nil {
		c.EmbeddedNats.Routes = []string{fmt.Sprintf("nats://%s.%s.internal:5222", region, appName)}
	}

	if rpcName, ok := os.LookupEnv("ANYCABLE_FLY_RPC_APP_NAME"); ok {
		if c.RPC.Host == defaults.RPC.Host {
			c.RPC.Host = fmt.Sprintf("dns:///%s.%s.internal:50051", region, rpcName)
		}
	}

	return nil
}

func (c *Config) loadHerokuPreset(defaults *Config) error {
	if c.Host == defaults.Host {
		c.Host = "0.0.0.0"
	}

	return nil
}

func detectPresetsFromEnv() []string {
	presets := []string{}

	if isFlyEnv() {
		presets = append(presets, "fly")
	}

	if isHerokuEnv() {
		presets = append(presets, "heroku")
	}

	return presets
}

func isFlyEnv() bool {
	if _, ok := os.LookupEnv("FLY_APP_NAME"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("FLY_ALLOC_ID"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("FLY_REGION"); !ok {
		return false
	}

	return true
}

func isHerokuEnv() bool {
	if _, ok := os.LookupEnv("HEROKU_APP_ID"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("HEROKU_DYNO_ID"); !ok {
		return false
	}

	return true
}
