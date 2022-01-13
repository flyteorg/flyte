.. _deployment-troubleshooting:

###############
Troubleshooting
###############

***********************
Cloudflare DNS provider
***********************

If when using Cloudflare as a DNS, you find that the browser works but ``flytectl`` does not, make sure your
proxy settings have been turned off. They should have been turned off automatically when you copied the nameservers from the output of ``opta output  -c env.yaml`` if you're using ``opta``. If proxying is turned on, Cloudflare or another DNS provider may filter out grpc traffic, which is what ``flytectl`` relies on.
