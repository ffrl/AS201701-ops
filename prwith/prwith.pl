#!/usr/bin/perl

# Author: takt
# Purpose: Find peerings of a certain AS with AS201701
# Version: 1.0

use warnings;
use strict;
use Data::Dumper;

my $asn = $ARGV[0];
if (!defined($asn)) {
        print STDERR "Usage: prwith <ASN>\n";
        exit;
}

my @routers = (
        'bb-a.fra3.fra.de.ffrl.de',
        'bb-b.fra3.fra.de.ffrl.de',
        'bb-a.ix.dus.de.ffrl.de',
        'bb-b.ix.dus.de.ffrl.de',
        'bb-a.ak.ber.de.ffrl.de',
        'bb-b.ak.ber.de.ffrl.de'
);

printf("%-25s%-25s%-10s%-25s%-25s%-25s%-25s\n", "Router", "Neighbor", "AS", "Filter In", "Filter Out", "Status", "Routes");
foreach my $router (@routers) {
        get_peerings($router, $asn);
}

sub get_peerings {
        my ($router, $asn) = @_;

        my $remote_cmd = "sudo birdc \\\"show protocol all\\\"";
        my $remote_cmd6 = "sudo birdc6 \\\"show protocol all\\\"";
        my $local_cmd = "ssh -p 2222 $router \"$remote_cmd\" 2>/dev/null";
        my $local_cmd6 = "ssh -p 2222 $router \"$remote_cmd6\" 2>/dev/null";

        my @output4 = `$local_cmd`;
        my @output6 = `$local_cmd6`;
        my @output = (@output4, @output6);

        my $i = 0;

        my @peerings = extract_peerings(@output);

        foreach my $peering (@peerings) {
                if ($peering->{'asn'} != $asn) {
                        next;
                }
                printf("%-25s%-25s%-10s%-25s%-25s%-25s%-10s%-10s\n", $router, $peering->{'neighbor'}, $peering->{'asn'}, $peering->{'filter_in'}, $peering->{'filter_out'}, $peering->{'state'}, $peering->{'routes_rec'}, $peering->{'routes_send'});
        }
}

sub extract_peerings {
        my (@lines) = @_;
        my @peerings;
        my $i = 0;
        my $in_bgp_block = 0;

        my $bgp_status;
        my $bgp_filter_in;
        my $bgp_filter_out;
        my $bgp_asn;
        my $bgp_neighbor;
        my $bgp_received;
        my $bgp_announced;
        my $tmp;
        my $peer;

        foreach my $line (@lines) {
                if ($i <= 1) {
                        $i++;
                        next;
                }

                chomp($line);
                if ($in_bgp_block == 0) {
                        if ($line =~ /BGP.*master/) {
                                $in_bgp_block = 1;
                                ($tmp, $tmp, $tmp, $tmp, $tmp, $bgp_status) = split(/\s+/, $line);
                        }
                }



                if ($in_bgp_block == 1) {
                        if ($line =~ /Neighbor address/) {
                                ($tmp, $tmp, $tmp, $bgp_neighbor) = split(/\s+/, $line);
                                next;
                        }

                        if ($line =~ /Neighbor AS/) {
                                ($tmp, $tmp, $tmp, $bgp_asn) = split(/\s+/, $line);
                                next;
                        }

                        if ($line =~ /Input filter/) {
                                ($tmp, $tmp, $tmp, $bgp_filter_in) = split(/\s+/, $line);
                                next;
                        }

                        if ($line =~ /Output filter/) {
                                ($tmp, $tmp, $tmp, $bgp_filter_out) = split(/\s+/, $line);
                                next;
                        }

                        if ($line =~ /Routes\:/) {
                                ($tmp, $tmp, $bgp_received, $tmp, $bgp_announced) = split(/\s+/, $line);
                        }

                        if ($line =~ /^$/) {
                                $in_bgp_block = 0;

                                $peer = {
                                        'neighbor' => $bgp_neighbor,
                                        'asn' => $bgp_asn,
                                        'state' => $bgp_status,
                                        'filter_in' => $bgp_filter_in,
                                        'filter_out' => $bgp_filter_out,
                                        'routes_rec' => $bgp_received,
                                        'routes_send' => $bgp_announced
                                };

                                push(@peerings, $peer);
                        }
                }

        }
        return @peerings;
}
