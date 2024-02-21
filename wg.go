package WG

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bepass-org/wireguard-go/app"

	_ "golang.org/x/mobile/bind"
)

type Flags struct {
	Verbose        bool
	BindAddress    string
	Endpoint       string
	License        string
	Country        string
	PsiphonEnabled bool
	Gool           bool
	Scan           bool
	Rtt            int
}

var validFlags = map[string]bool{
	"-v":       true,
	"-b":       true,
	"-e":       true,
	"-k":       true,
	"-country": true,
	"-cfon":    true,
	"-gool":    true,
	"-scan":    true,
	"-rtt":     true,
}

func newFlags() *Flags {
	return &Flags{}
}

func (f *Flags) setup() {
	flag.BoolVar(&f.Verbose, "v", false, "verbose")
	flag.StringVar(&f.BindAddress, "b", "127.0.0.1:8086", "socks bind address")
	flag.StringVar(&f.Endpoint, "e", "notset", "warp clean IP")
	flag.StringVar(&f.License, "k", "notset", "license key")
	flag.StringVar(&f.Country, "country", "", "psiphon country code in ISO 3166-1 alpha-2 format")
	flag.BoolVar(&f.PsiphonEnabled, "cfon", false, "enable Psiphon over warp")
	flag.BoolVar(&f.Gool, "gool", false, "enable warp gooling")
	flag.BoolVar(&f.Scan, "scan", false, "enable warp scanner(experimental)")
	flag.IntVar(&f.Rtt, "rtt", 1000, "scanner rtt threshold, default 1000")

	flag.Usage = usage
	flag.Parse()
}

var validCountryCodes = map[string]bool{
	"AT": true,
	"BE": true,
	"BG": true,
	"BR": true,
	"CA": true,
	"CH": true,
	"CZ": true,
	"DE": true,
	"DK": true,
	"EE": true,
	"ES": true,
	"FI": true,
	"FR": true,
	"GB": true,
	"HU": true,
	"IE": true,
	"IN": true,
	"IT": true,
	"JP": true,
	"LV": true,
	"NL": true,
	"NO": true,
	"PL": true,
	"RO": true,
	"RS": true,
	"SE": true,
	"SG": true,
	"SK": true,
	"UA": true,
	"US": true,
}

func usage() {
	log.Println("./warp-plus-go [-v] [-b addr:port] [-c config-file-path] [-e warp-ip] [-k license-key] [-country country-code] [-cfon] [-gool]")
	flag.PrintDefaults()
}

func validateFlags(f *Flags) error {
	if _, err := net.ResolveTCPAddr("tcp", f.BindAddress); err != nil {
		return fmt.Errorf("invalid bindAddress format: %s", f.BindAddress)
	}

	if ip := net.ParseIP(f.Endpoint); ip == nil {
		return fmt.Errorf("invalid warp clean IP: %s", f.Endpoint)
	}

	if f.PsiphonEnabled && f.Country == "" {
		return fmt.Errorf("if Psiphon is enabled, country code must be provided")
	}

	if f.PsiphonEnabled && !validCountryCodes[f.Country] {
		validCountries := make([]string, 0, len(validCountryCodes))

		for code, _ := range validCountryCodes {
			validCountries = append(validCountries, code)
		}

		return fmt.Errorf("invalid country code: %s. valid country codes: %s", f.Country, validCountries)
	}

	return nil
}

func parseCommandLine(command string) ([]string, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, fmt.Errorf("unclosed quote in command line: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}

type Instance struct {
	done chan bool
	args string
}

func NewInstance(args string) *Instance {
	i := Instance{
		done: make(chan bool),
		args: args,
	}
	return &i
}

func (i *Instance) Quit() {
	i.done <- true
}

func (i *Instance) Run() error {
	argsNew, err := parseCommandLine("run " + i.args)

	// Check for unexpected flags
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}
	os.Args = argsNew

	flags := newFlags()
	flags.setup()

	if err := validateFlags(flags); err != nil {
		return fmt.Errorf("validatrion error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-i.done
		cancel()
	}()

	err = app.RunWarp(flags.PsiphonEnabled, flags.Gool, flags.Scan, flags.Verbose, flags.Country, flags.BindAddress, flags.Endpoint, flags.License, ctx, flags.Rtt)
	if err != nil {
		return err
	}

	return nil
}
