
using Printf

function calculate_area(a, b)
    π * a * b
end

function read_input(input_dir, v)
    open(@sprintf "%s/%s" input_dir v) do file
        parse.(Float64, read(file, String))
    end
end

function write_output(output_dir, output_file, v)
    output_path = @sprintf "%s/%s" output_dir output_file
    open(output_path, "w") do file
        write(file, string(v))
    end
end

function main(input_dir, output_dir)
    a = read_input(input_dir, 'a')
    b = read_input(input_dir, 'b')

    area = calculate_area(a, b)

    write_output(output_dir, "area", area)
    write_output(output_dir, "metadata", "[from julia rawcontainer]")
end

# the keyword ARGS is a special value that contains the command-line arguments
# julia arrays are 1-indexed
input_dir = ARGS[1]
output_dir = ARGS[2]

main(input_dir, output_dir)
