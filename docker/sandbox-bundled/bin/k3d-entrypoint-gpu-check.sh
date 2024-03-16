#!/bin/sh

if [ -n "${FLYTE_GPU}"  ]; then
  echo "GPU Enabled - checking if it's available"
  nvidia-smi
  if [ $? -eq 0 ]; then
    echo "nvidia-smi working"
  else
    >&2 echo "NVIDIA not available, enable it in docker like so: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/user-guide.html"
    exit 255 
  fi

else
  echo "GPU not enabled"
fi
