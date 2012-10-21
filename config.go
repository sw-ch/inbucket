package inbucket

import (
	"container/list"
	"fmt"
	"github.com/robfig/goconfig/config"
	"net"
	"os"
)

// SmtpConfig houses the SMTP server configuration - not using pointers
// so that I can pass around copies of the object safely.
type SmtpConfig struct {
	Ip4address net.IP
	Ip4port    int
	Domain     string
}

var smtpConfig *SmtpConfig

var Config *config.Config

// GetSmtpConfig returns a copy of the SmtpConfig
func GetSmtpConfig() SmtpConfig {
	return *smtpConfig
}

// LoadConfig loads the specified configuration file into inbucket.Config
// and performs validations on it.
func LoadConfig(filename string) error {
	var err error
	Config, err = config.ReadDefault(filename)
	if err != nil {
		return err
	}

	messages := list.New()

	// Validate sections
	requireSection(messages, "smtp")
	requireSection(messages, "web")
	requireSection(messages, "datastore")
	if messages.Len() > 0 {
		fmt.Fprintln(os.Stderr, "Error(s) validating configuration:")
		for e := messages.Front(); e != nil; e = e.Next() {
			fmt.Fprintln(os.Stderr, " -", e.Value.(string))
		}
		return fmt.Errorf("Failed to validate configuration")
	}

	// Validate options
	requireOption(messages, "smtp", "ip4.address")
	requireOption(messages, "smtp", "ip4.port")
	requireOption(messages, "smtp", "domain")
	requireOption(messages, "web", "ip4.address")
	requireOption(messages, "web", "ip4.port")
	requireOption(messages, "web", "templates.dir")
	requireOption(messages, "web", "public.dir")
	requireOption(messages, "datastore", "path")
	if messages.Len() > 0 {
		fmt.Fprintln(os.Stderr, "Error(s) validating configuration:")
		for e := messages.Front(); e != nil; e = e.Next() {
			fmt.Fprintln(os.Stderr, " -", e.Value.(string))
		}
		return fmt.Errorf("Failed to validate configuration")
	}

	err = parseSmtpConfig()

	return err
}

// parseSmtpConfig trying to catch config errors early
func parseSmtpConfig() error {
	smtpConfig = new(SmtpConfig)

	// Parse IP4 address only, error on IP6.
	option := "[smtp]ip4.address"
	str, err := Config.String("smtp", "ip4.address")
	if err != nil {
		return fmt.Errorf("Failed to parse %v: %v", option, err)
	}
	addr := net.ParseIP(str)
	if addr == nil {
		return fmt.Errorf("Failed to parse %v '%v'", option, str)
	}
	addr = addr.To4()
	if addr == nil {
		return fmt.Errorf("Failed to parse %v '%v' not IPv4!", option, str)
	}
	smtpConfig.Ip4address = addr

	option = "[smtp]ip4.port"
	smtpConfig.Ip4port, err = Config.Int("smtp", "ip4.port")
	if err != nil {
		return fmt.Errorf("Failed to parse %v: %v", option, err)
	}

	option = "[smtp]domain"
	str, err = Config.String("smtp", "domain")
	if err != nil {
		return fmt.Errorf("Failed to parse %v: %v", option, err)
	}
	smtpConfig.Domain = str

	return nil
}

// requireSection checks that a [section] is defined in the configuration file,
// appending a message if not.
func requireSection(messages *list.List, section string) {
	if !Config.HasSection(section) {
		messages.PushBack(fmt.Sprintf("Config section [%v] is required", section))
	}
}

// requireOption checks that 'option' is defined in [section] of the config file,
// appending a message if not.
func requireOption(messages *list.List, section string, option string) {
	if !Config.HasOption(section, option) {
		messages.PushBack(fmt.Sprintf("Config option '%v' is required in section [%v]", option, section))
	}
}
