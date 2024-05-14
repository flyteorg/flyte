import System.IO
import System.Environment
import Text.Read
import Text.Printf

calculateEllipseArea :: Float -> Float -> Float
calculateEllipseArea a b = pi * a * b

main = do
  args <- getArgs
  let a = args!!0
      b = args!!1

  let area = calculateEllipseArea (read a::Float) (read b::Float)

  let output_area = args!!2 ++ "/area"
      output_metadata = args!!2 ++ "/metadata"
  writeFile output_area (show area)
  writeFile output_metadata "[from haskell rawcontainer]"
