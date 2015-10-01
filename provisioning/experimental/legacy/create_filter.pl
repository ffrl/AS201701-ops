#!/usr/bin/perl

use warnings;
use strict;

if (scalar(@ARGV) < 3) {
        print STDERR "Usage: ./create_filter.pl <COMMUNITY> <IPV4> <IPV6>";
}

my $community = $ARGV[0];
my $ipv4 = $ARGV[1];
my $ipv6 = $ARGV[2];

print "ip prefix-list ipv4-$community-in seq 5 permit $ipv4 le 32\n";
print "ip prefix-list ipv4-$community-in seq 10 deny any\n";

print "ipv6 prefix-list ipv6-$community-in seq 5 permit $ipv6 le 56\n";
print "ipv6 prefix-list ipv6-$community-in seq 10 deny any\n";
