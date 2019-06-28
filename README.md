# crosswords

i wrote some code to generate valid NYT crossword grids, and it started getting a little out of hand

this code is disgusting but it can 
 - generate all valid crosswords of a given size and
 - count the number of (2n + 1) x (2n + 1) crosswords for using a DP (it matches https://oeis.org/A323838 through 15x15 so pretty confident it's correct!)
 - i also verified some other sequences, and added 23x23 mirror-symmetric counts to https://oeis.org/A325408 and https://oeis.org/A325409!
