#!/usr/bin/env perl
use warnings;
use strict;

print "module Cardea\n";

while (<>) {
  /^\s*var\s+(\w+)\s*=\s*regexp\.MustCompile\(\"(.*)\"\)$/ or next;
  my ($name, $rx) = ($1, $2);
  $rx =~ s/\(\?P</(?</g;
  print "  $name = Regexp.compile(\"$rx\");\n";
}

print "end\n";
