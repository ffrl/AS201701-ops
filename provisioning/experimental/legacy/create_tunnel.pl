#!/usr/bin/perl

use warnings;
use strict;

my %routers = (
        'bb-a.ix.dus.de.ffrl.de' => "185.66.193.0",
        'bb-b.ix.dus.de.ffrl.de' => "185.66.193.1",
        'bb-a.fra3.fra.de.ffrl.de' => "185.66.194.0",
        'bb-b.fra3.fra.de.ffrl.de' => "185.66.194.1",
        'bb-b.ak.ber.de.ffrl.de' => "185.66.195.1",
);

if (scalar(@ARGV) < 7) {
        print STDERR "Usage: ./create_tunnel.pl <INT-NAME> <TUN-A> <TUN-B> <TUN-IPV4> <TUN-IPV6> <COMMUNITY> <DESCRIPTION> <ASN>";
}

my $tunnel_name = $ARGV[0];
my $tunnel_endpoint_a = $ARGV[1];
my $tunnel_endpoint_b = $ARGV[2];
my $tunnel_ipv4_a = "$ARGV[3]";
my ($tmp1, $tmp2, $tmp3, $last) = split(/\./, $ARGV[3]);
$last++;
my $tunnel_ipv4_b = "$tmp1.$tmp2.$tmp3.$last";
my $tunnel_ipv6_a = "$ARGV[4]1";
my $tunnel_ipv6_b = "$ARGV[4]2";
my $community = $ARGV[5];
my $bgp_description = $ARGV[6];
my $asn = $ARGV[7];

my $a = $routers{$tunnel_endpoint_a};
my $file = "Description='GRE tunnel $tunnel_name'
Interface=$tunnel_name
Connection=tunnel
Mode='gre'
Local='$a'
Remote='$tunnel_endpoint_b'
IP=static
Address=$tunnel_ipv4_a/31
ExecUpPost='ip tunnel change $tunnel_name ttl 64 && ip link set mtu 1400 dev $tunnel_name'

IP6=static
Address6='$tunnel_ipv6_a/64'";

# Write config file:
open(my $fh, ">", "/tmp/$tunnel_name");
print $fh $file;
close($fh);

# Copy file to router
my $cmd = "scp -P 2222 /tmp/$tunnel_name takt\@$a:/home/takt";
#`$cmd`;

$cmd = "ssh -p 2222 takt\@$a \"sudo mv /home/takt/$tunnel_name /etc/netctl/$tunnel_name\" 2>/dev/null";
#`$cmd`;

$cmd = "ssh -p 2222 takt\@$a \"sudo netctl stop $tunnel_name\" 2>/dev/null";
#`$cmd`;

$cmd = "ssh -p 2222 takt\@$a \"sudo netctl start $tunnel_name\" 2>/dev/null";
#`$cmd`;

print "router bgp 201701\n";
print "neighbor $tunnel_ipv4_b remote-as $asn\n";
print "neighbor $tunnel_ipv4_b peer-group ipv4-community\n";
print "neighbor $tunnel_ipv4_b description $bgp_description\n";
print "neighbor $tunnel_ipv4_b prefix-list ipv4-ffrl-$community-in in\n";
print "neighbor $tunnel_ipv6_b remote-as $asn\n";
print "no neighbor $tunnel_ipv6_b activate\n";
print "address-family ipv6\n";
print "neighbor $tunnel_ipv6_b peer-group ipv6-community\n";
print "neighbor $tunnel_ipv6_b prefix-list ipv6-ffrl-$community-in in\n";
print "exit\n";
print "exit\n";
