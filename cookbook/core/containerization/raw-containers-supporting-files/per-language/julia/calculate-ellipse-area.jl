
using Printf

function calculate_area(a, b)
    π * a * b
end

function write_output(output_dir, output_file, v)
    output_path = @sprintf "%s/%s" output_dir output_file
    open(output_path, "w") do file
        write(file, string(v))
    end
end

function main(a, b, output_dir)
    a = parse.(Float64, a)
    b = parse.(Float64, b)

    area = calculate_area(a, b)

    write_output(output_dir, "area", area)
    write_output(output_dir, "metadata", "[from julia rawcontainer]")
end

# the keyword ARGS is a special value that contains the command-line arguments
# julia arrays are 1-indexed
a = ARGS[1]
b = ARGS[2]
output_dir = ARGS[3]

main(a, b, output_dir)
