FROM r-base

WORKDIR /root

COPY *.R /root/

# Not sure whether this a horrible hack. I couldn't
# find a better way to install a package via command
# line
RUN Rscript --save install-readr.R
