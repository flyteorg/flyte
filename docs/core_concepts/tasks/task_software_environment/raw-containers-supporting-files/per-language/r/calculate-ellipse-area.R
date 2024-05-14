#!/usr/bin/env Rscript

args = commandArgs(trailingOnly=TRUE)

a = args[1]
b = args[2]
output_dir = args[3]

area <- pi * as.double(a) * as.double(b)
print(area)

writeLines(as.character(area), sprintf("%s/%s", output_dir, 'area'))
writeLines("[from R rawcontainer]", sprintf("%s/%s", output_dir, 'metadata'))
