/*
 * smasher.go is a tool that audits the running configuration of
 * Linux backbone routers within AS-FFRL / AS201701
 *
 * (c) 2015 by:
 *  takt
 */

package main

import "fmt"
import "os"
import "io/ioutil"
import "os/exec"
import "strings"
import "regexp"
import "strconv"
import "flag"

const VERSION = "0.9"
const PROC_SYS_NET_IPV4 = "/proc/sys/net/ipv4"
const PROC_SYS_NET_IPV6 = "/proc/sys/net/ipv6"
const BIRDC = "/usr/sbin/birdc"
const BIRDC6 = "/usr/sbin/birdc6"
const BIRD_PROTOCOL_STATUS_COLUMN = 5
const BIRD_PROTOCOL_NAME_COLUMN = 0
const BIRD_PROTOCOL_ROUTES = 6
const BIRD_PROTOCOL_ROUTES_EXPORTED = 4
const IP = "/sbin/ip"
const IPV4_FULL_TABLE = 560000
const IPV6_FULL_TABLE = 25000
const RIB_FIB_TOLERANCE = 0.02
const BIRD_CONF = "/etc/bird/bird.conf"
const BIRD6_CONF = "/etc/bird/bird6.conf"

var show_version = flag.Bool("version", false, "Show software version")

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func check_ip_forwarding(interface_name string) {
	ipv4_forwarding, err := ioutil.ReadFile(PROC_SYS_NET_IPV4 + "/conf/" +
	  interface_name + "/forwarding")
	check(err)

	if string(ipv4_forwarding) != "1\n" {
		fmt.Printf("IPv4 forwarding on %s: disabled\n", interface_name)
	}

	ipv6_forwarding, err := ioutil.ReadFile(PROC_SYS_NET_IPV6 + "/conf/" +
	  interface_name + "/forwarding")
	check(err)

	if string(ipv6_forwarding) != "1\n" {
		fmt.Printf("IPv6 forwarding on %s: disabled\n", interface_name)
	}
}

func check_rp_filter(interface_name string) {
	rp_filter, err := ioutil.ReadFile(PROC_SYS_NET_IPV4 + "/conf/" +
	  interface_name + "/rp_filter")
	check(err)

	if string(rp_filter) != "0\n" {
		fmt.Printf("rp_filter on %s: enabled\n", interface_name)
	}
}

func check_interfaces() {
	interfaces, err := ioutil.ReadDir(PROC_SYS_NET_IPV4 + "/conf")
	check(err)

	for _, interface_ := range interfaces {
		check_ip_forwarding(interface_.Name())
		check_rp_filter(interface_.Name())
		/*
		 * TODO: Check IP addresses on interfaces against IPDB.
		 */
	}
}

func check_ibgp(version int) {
	var cmd *exec.Cmd
	if version == 4 {
		cmd = exec.Command(BIRDC, "show protocols")
	} else if version == 6 {
		cmd = exec.Command(BIRDC6, "show protocols")
	} else {
		return;
	}
	output, err := cmd.CombinedOutput()
	check(err)

	lines := strings.Split(string(output), "\n")
	for i := 0; i < len(lines); i++ {
		if i < 2 {
			continue
		}

		matched, err := regexp.MatchString("^bb_[ab]", lines[i])
		check(err)
		if !matched {
			continue
		}

		columns := regexp.MustCompile(" +").Split(string(lines[i]), -1)
		if columns[BIRD_PROTOCOL_STATUS_COLUMN] != "Established" {
			fmt.Printf("IPv%d IBGP session %s is in status %s\n", version,
			  columns[BIRD_PROTOCOL_NAME_COLUMN],
			  columns[BIRD_PROTOCOL_STATUS_COLUMN])
		}
	}
}

func check_routes(version int) {
	var cmd *exec.Cmd
	if version == 4 {
		cmd = exec.Command(BIRDC, "show protocols all kernel1")
	} else if version == 6 {
		cmd = exec.Command(BIRDC6, "show protocols all kernel1")
	} else {
		return;
	}
	output, err := cmd.CombinedOutput()
	check(err)

	lines := strings.Split(string(output), "\n")
	columns := regexp.MustCompile(" +").Split(string(lines[BIRD_PROTOCOL_ROUTES]), -1)
	rib_routes, err := strconv.ParseInt(columns[BIRD_PROTOCOL_ROUTES_EXPORTED], 10, 32)

	if version == 4 {
		cmd = exec.Command(IP, "route", "show")
	} else if version == 6 {
		cmd = exec.Command(IP, "-6", "route", "show")
	}
	output, err = cmd.CombinedOutput()
	check(err)

	lines = strings.Split(string(output), "\n")

	var fib_routes int = 0
	for i := 0; i < len(lines); i++ {
		matched, err := regexp.MatchString("proto bird", lines[i])
		check(err)
		if matched {
			fib_routes++
		}
	}

	if version == 4 {
		if  rib_routes < IPV4_FULL_TABLE {
			fmt.Printf("IPv4 RIB < %d routes: %d\n", IPV4_FULL_TABLE, rib_routes)
		}
	}
	if version == 6 {
		if int(rib_routes) < IPV6_FULL_TABLE {
			fmt.Printf("IPv6 RIB < %d routes: %d\n", IPV6_FULL_TABLE, rib_routes)
		}
	}

	if float64(fib_routes) < float64(rib_routes) * (float64(1) - float64(RIB_FIB_TOLERANCE)) {
		fmt.Printf("IPv%d RIB and FIB differing by more than %f\n", version,
		  RIB_FIB_TOLERANCE)
		fmt.Printf("IPv%d RIB: %d FIB: %d\n", version, rib_routes, fib_routes)
	}
}

func check_router_drain(version int) {
	var filename string
	if version == 4 {
		filename = BIRD_CONF
	}
	if version == 6 {
		filename = BIRD6_CONF
	}

	output, err := ioutil.ReadFile(filename)
	check(err)

	lines := strings.Split(string(output), "\n")
	for i := 0; i < len(lines); i++ {
		line := string(lines[i])
		matched, err := regexp.MatchString("define DRAINED", line)
		check(err)
		if matched {
			if line == "define DRAINED = 1;" {
				fmt.Printf("Router is IPv%d drained\n", version)
			}
		}
	}
}

func check_router_metroization(version int) {
	var filename string
	if version == 4 {
		filename = BIRD_CONF
	}
	if version == 6 {
		filename = BIRD6_CONF
	}

	output, err := ioutil.ReadFile(filename)
	check(err)

	lines := strings.Split(string(output), "\n")
	for i := 0; i < len(lines); i++ {
		line := string(lines[i])
		matched, err := regexp.MatchString("define METROIZED", line)
		check(err)
		if matched {
			if line == "define METROIZED = 1;" {
				fmt.Printf("Router is IPv%d metroized\n", version)
			}
		}
	}
}

func main() {
	flag.Parse()
	if *show_version == true {
		fmt.Printf("smasher version %s\n", VERSION)
		os.Exit(0)
	}

	check_interfaces()
	check_ibgp(4)
	check_ibgp(6)
	check_routes(4)
	check_routes(6)
	check_router_drain(4)
	check_router_drain(6)
	check_router_metroization(4)
	check_router_metroization(6)
	/*
	 * TODO:
	 *  check OSPF status
	 */

	os.Exit(0)
}
