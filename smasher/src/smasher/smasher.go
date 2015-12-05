/*
 * smasher.go is a tool that audits the running configuration of
 * routers within AS-FFRL / AS201701
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

const PROC_SYS_NET_IPV4 = "/proc/sys/net/ipv4"
const PROC_SYS_NET_IPV6 = "/proc/sys/net/ipv6"
const BIRDC = "/usr/sbin/birdc"
const BIRDC6 = "/usr/sbin/birdc6"
const BIRD_PROTOCOL_STATUS_COLUMN = 5
const BIRD_PROTOCOL_NAME_COLUMN = 0

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
			  columns[BIRD_PROTOCOL_NAME_COLUMN], columns[BIRD_PROTOCOL_STATUS_COLUMN])
		}
	}
}

func main() {
	check_interfaces()
	check_ibgp(4)
	check_ibgp(6)

	os.Exit(0)
}
