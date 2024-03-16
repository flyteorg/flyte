#!/bin/bash

set -ex

cat "{inputs.f}" >> "{outputs.z}"
echo "Hello World {inputs.y} on  {inputs.j} - output {outputs.x}"
