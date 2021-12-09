library(readr)

args = commandArgs(trailingOnly=TRUE)

input_dir = args[1]
output_dir = args[2]

a = read_lines(sprintf("%s/%s", input_dir, 'a'))
b = read_lines(sprintf("%s/%s", input_dir, 'b'))

area <- pi * as.double(a) * as.double(b)
print(area)

writeLines(as.character(area), sprintf("%s/%s", output_dir, 'area'))
writeLines("[from R rawcontainer]", sprintf("%s/%s", output_dir, 'metadata'))
