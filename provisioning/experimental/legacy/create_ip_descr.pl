#!/usr/bin/perl

use warnings;
use strict;

my @routers = ("bb-a.fra3.fra.de.ffrl.de", "bb-b.fra3.fra.de.ffrl.de", "bb-a.ix.dus.de.ffrl.de", "bb-b.ix.dus.de.ffrl.de", "bb-b.ak.ber.de.ffrl.de");
my %communities = (
        'bucht' => {
                'asn' => 64859,
                'supernodes' => {
                        'bucht1' => '89.163.150.92',
                        'bucht2' => '89.163.150.93'
                }
        },
        'neuss' => {
                'asn' => 64857,
                'supernodes' => {
                        'neuss1' => '89.163.154.94',
                        'neuss2' => '89.163.154.95'
                }
        },
        'ddorf' => {
                'asn' => 64856,
                'supernodes' => {
                        'ddorf1' => '89.163.150.94',
                        'ddorf2' => '89.163.154.93'
                }
        },
        'tiefl' => {
                'asn' => 64858,
                'supernodes' => {
                        'tiefl1' => '89.163.154.96',
                        'tiefl2' => '89.163.154.97'
                }
        }
);

my $start_addr_v4 = 40;
my $start_addr_v6 = 0x9b;

my $a = $start_addr_v4;
my $b = $start_addr_v6;

foreach my $router (@routers) {
        #print "CONFIG FOR: $router\n";
        foreach my $community (sort keys %communities) {
                foreach my $supernode (sort keys $communities{$community}->{'supernodes'}) {
                        #print $supernode . "\n";

                        my $b_hex = sprintf("%x", $b);
                        #my $cmd = "./create_tunnel.pl t-ffrl-$supernode $router $communities{$community}->{'supernodes'}->{$supernode} 100.64.1.$a 2a03:2260:0:". $b_hex .":: $community $supernode $communities{$community}->{'asn'}\n";
                        #print `$cmd`;

                        print "100.64.1.$a/31 $router <> $supernode\n";
                        print "2a03:2260:0:" . $b_hex . "::/64 $router <> $supernode\n";

                        $a += 2;
                        $b++;

                }

        }
}
