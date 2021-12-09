import System.IO
import System.Environment
import Text.Read
import Text.Printf

calculateEllipseArea :: Float -> Float -> Float
calculateEllipseArea a b = pi * a * b

main = do
  args <- getArgs
  let input_a = args!!0 ++ "/a"
      input_b = args!!0 ++ "/b"
  a <- readFile input_a
  b <- readFile input_b

  let area = calculateEllipseArea (read a::Float) (read b::Float)

  let output_area = args!!1 ++ "/area"
      output_metadata = args!!1 ++ "/metadata"
  writeFile output_area (show area)
  writeFile output_metadata "[from haskell rawcontainer]"
