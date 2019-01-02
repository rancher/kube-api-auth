#!/bin/bash
<glide.yaml head -1 | sed 's/^package: //'
echo
echo '# generated from glide.lock'
<glide.lock grep -A1 '^- name:' | perl -pe '
    s/^--\n$//;
    if (s/^- name: (.*)\n/$1/) {
        $count = 64-length($1);
        $space = " "x$count;
        s/$/$space/; 
    }
    s/ version: //;
'
